package null

import (
	"context"
	"net/url"
	"strings"

	"github.com/streamingfast/dauth"
	"go.uber.org/zap"
)

func Register() {
	dauth.Register("null", func(configURL string, _ *zap.Logger) (dauth.Authenticator, error) {
		headers := map[string]string{}

		urlObject, err := url.Parse(configURL)
		if err == nil {
			params := urlObject.Query()

			for key, values := range params {
				switch key {
				case "user_id":
					headers[dauth.SFHeaderUserID] = values[0]
				case "api_key_id":
					headers[dauth.SFHeaderApiKeyID] = values[0]
				case "meta":
					headers[dauth.SFHeaderMeta] = values[0]
				default:
					headers[key] = strings.Join(values, ",")
				}
			}
		}

		return &nullPlugin{
			trustedHeaders: headers,
		}, nil
	})
}

type nullPlugin struct {
	trustedHeaders map[string]string
}

func (t *nullPlugin) Ready(_ context.Context) bool {
	return true
}

func (t *nullPlugin) Authenticate(ctx context.Context, _ string, _ map[string][]string, ipAddress string) (context.Context, error) {
	out := make(dauth.TrustedHeaders)
	out[dauth.SFHeaderIP] = ipAddress

	for key, value := range t.trustedHeaders {
		out[key] = value
	}

	return dauth.WithTrustedHeaders(ctx, out), nil
}
