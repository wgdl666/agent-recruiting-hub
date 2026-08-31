package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	modelhubv2 "github.com/wgdl666/wgModelHub/gen/wg_model_hub/v2"
	"github.com/wgdl666/wgModelHub/protocol"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	defaultModelHubModel = "gemini-2.5-flash"
	defaultCallerService = "agent-recruiting-hub"
)

// ModelHub calls wgModelHub gRPC Generate for text JSON scoring.
type ModelHub struct {
	Address string
	Model   string
	Caller  string
	Timeout time.Duration

	once sync.Once
	conn *grpc.ClientConn
	rpc  modelhubv2.ModelHubServiceClient
	dial error
}

func NewFromEnv() *ModelHub {
	addr := strings.TrimSpace(os.Getenv("HUB_MODELHUB_ADDRESS"))
	if addr == "" {
		addr = strings.TrimSpace(os.Getenv("MODEL_HUB_ADDRESS"))
	}
	if addr == "" {
		return nil
	}
	model := strings.TrimSpace(os.Getenv("HUB_MODELHUB_MODEL"))
	if model == "" {
		model = strings.TrimSpace(os.Getenv("HUB_GEMINI_MODEL"))
	}
	if model == "" {
		model = defaultModelHubModel
	}
	caller := strings.TrimSpace(os.Getenv("HUB_MODELHUB_CALLER"))
	if caller == "" {
		caller = defaultCallerService
	}
	return &ModelHub{
		Address: addr,
		Model:   model,
		Caller:  caller,
		Timeout: 90 * time.Second,
	}
}

func (c *ModelHub) Enabled() bool {
	return c != nil && strings.TrimSpace(c.Address) != ""
}

func (c *ModelHub) ModelName() string {
	if c == nil || c.Model == "" {
		return defaultModelHubModel
	}
	return c.Model
}

func (c *ModelHub) ensure() error {
	if c == nil {
		return fmt.Errorf("modelhub not configured")
	}
	c.once.Do(func() {
		conn, err := grpc.NewClient(
			c.Address,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithUnaryInterceptor(unaryCallerInterceptor(c.Caller)),
			grpc.WithStreamInterceptor(streamCallerInterceptor(c.Caller)),
			grpc.WithDefaultCallOptions(
				grpc.MaxCallRecvMsgSize(protocol.MaxRPCMessageBytes),
				grpc.MaxCallSendMsgSize(protocol.MaxRPCMessageBytes),
			),
		)
		if err != nil {
			c.dial = fmt.Errorf("dial modelhub %s: %w", c.Address, err)
			return
		}
		c.conn = conn
		c.rpc = modelhubv2.NewModelHubServiceClient(conn)
	})
	return c.dial
}

func unaryCallerInterceptor(caller string) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		return invoker(withCallerMetadata(ctx, caller), method, req, reply, cc, opts...)
	}
}

func streamCallerInterceptor(caller string) grpc.StreamClientInterceptor {
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		return streamer(withCallerMetadata(ctx, caller), desc, cc, method, opts...)
	}
}

func withCallerMetadata(ctx context.Context, caller string) context.Context {
	md, ok := metadata.FromOutgoingContext(ctx)
	if ok {
		md = md.Copy()
	} else {
		md = metadata.MD{}
	}
	md.Set(protocol.CallerMetadataKey, caller)
	return metadata.NewOutgoingContext(ctx, md)
}

// GenerateJSON sends system/user prompts and unmarshals JSON text output into dest.
func (c *ModelHub) GenerateJSON(system, user string, dest any) error {
	return c.GenerateJSONOpts(system, user, dest, 4096, 0)
}

func (c *ModelHub) GenerateJSONOpts(system, user string, dest any, maxTokens int32, timeout time.Duration) error {
	return c.GenerateJSONParts(system, user, nil, dest, maxTokens, timeout)
}

// UserMedia 评分时把图片 PDF / 页面截图一并交给模型，避免纯文字层抽不出来。
type UserMedia struct {
	MIME string
	Data []byte
}

func (c *ModelHub) GenerateJSONParts(system, user string, media []UserMedia, dest any, maxTokens int32, timeout time.Duration) error {
	if err := c.ensure(); err != nil {
		return err
	}
	system = strings.TrimSpace(system)
	user = strings.TrimSpace(user)
	if user == "" && len(media) == 0 {
		return fmt.Errorf("empty user prompt")
	}
	if maxTokens <= 0 {
		maxTokens = 4096
	}

	temp := 0.2
	textOut := &modelhubv2.TextOutput{
		Temperature:     &temp,
		MaxOutputTokens: &maxTokens,
		ResponseFormat: &modelhubv2.ResponseFormat{
			Type: modelhubv2.ResponseFormatType_RESPONSE_FORMAT_TYPE_JSON_OBJECT,
		},
		Thinking: modelhubv2.ThinkingMode_THINKING_MODE_DISABLED,
	}

	items := make([]*modelhubv2.InputItem, 0, 2)
	if system != "" {
		items = append(items, &modelhubv2.InputItem{Item: &modelhubv2.InputItem_Message{Message: &modelhubv2.Message{
			Role:  modelhubv2.Role_ROLE_SYSTEM,
			Parts: []*modelhubv2.ContentPart{{Content: &modelhubv2.ContentPart_Text{Text: system}}},
		}}})
	}
	userParts := make([]*modelhubv2.ContentPart, 0, 1+len(media))
	if user != "" {
		userParts = append(userParts, &modelhubv2.ContentPart{Content: &modelhubv2.ContentPart_Text{Text: user}})
	}
	for _, m := range media {
		if len(m.Data) == 0 || strings.TrimSpace(m.MIME) == "" {
			continue
		}
		blob := &modelhubv2.Media{MimeType: m.MIME, Source: &modelhubv2.Media_Data{Data: m.Data}}
		part := &modelhubv2.ContentPart{}
		if strings.HasPrefix(m.MIME, "image/") {
			part.Content = &modelhubv2.ContentPart_Image{Image: blob}
		} else {
			part.Content = &modelhubv2.ContentPart_File{File: blob}
		}
		userParts = append(userParts, part)
	}
	if len(userParts) == 0 {
		return fmt.Errorf("empty user prompt")
	}
	items = append(items, &modelhubv2.InputItem{Item: &modelhubv2.InputItem_Message{Message: &modelhubv2.Message{
		Role:  modelhubv2.Role_ROLE_USER,
		Parts: userParts,
	}}})

	wait := timeout
	if wait <= 0 {
		wait = c.Timeout
	}
	if wait <= 0 {
		wait = 90 * time.Second
	}
	ctx, cancel := context.WithTimeout(context.Background(), wait)
	defer cancel()

	stream, err := c.rpc.Generate(ctx, &modelhubv2.GenerateRequest{
		Model:  c.Model,
		Input:  &modelhubv2.Input{Items: items},
		Output: &modelhubv2.OutputSpec{Kind: &modelhubv2.OutputSpec_Text{Text: textOut}},
	})
	if err != nil {
		return formatRPCError("modelhub Generate", err)
	}

	var final *modelhubv2.GenerateEvent
	for {
		event, recvErr := stream.Recv()
		if recvErr == io.EOF {
			break
		}
		if recvErr != nil {
			return formatRPCError("modelhub Generate recv", recvErr)
		}
		if event.GetFinal() {
			final = event
		}
	}
	if final == nil {
		return fmt.Errorf("modelhub ended without final event")
	}
	if final.GetSafety().GetBlocked() {
		msg := strings.TrimSpace(final.GetSafety().GetMessage())
		if msg == "" {
			msg = "blocked"
		}
		return fmt.Errorf("modelhub blocked: %s", msg)
	}

	var text strings.Builder
	for _, item := range final.GetItems() {
		text.WriteString(item.GetText())
	}
	raw := strings.TrimSpace(text.String())
	if raw == "" {
		return fmt.Errorf("modelhub returned empty text")
	}
	if err := json.Unmarshal([]byte(raw), dest); err != nil {
		return fmt.Errorf("modelhub json: %w (body=%s)", err, truncate(raw, 200))
	}
	return nil
}

func formatRPCError(operation string, err error) error {
	if st, ok := status.FromError(err); ok && st.Message() != "" {
		return fmt.Errorf("%s: %s", operation, st.Message())
	}
	return fmt.Errorf("%s: %w", operation, err)
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
