package redisprefixtest

import (
	"context"
	"errors"
	"testing"

	"github.com/redis/go-redis/v9"
)

// PrefixHook is a hook that transparently prefixes each command with the name
// of the test it is instantiated in. It encapsulates all necessary data for
// modification, including the prefix string and the delimiter used to separate
// the namespace from the original command.
type PrefixHook struct {
	prefix    string
	delimiter string
}

// New returns a new PrefixHook instance intended for use in test environments
// only. By default, it prefixes commands with the test name obtained from
// tb.Name() and uses ":" as the delimiter. To customize the delimiter, provide
// alternative Options as needed.
func New(tb testing.TB, opts ...Option) *PrefixHook {
	options := options{
		prefix:    tb.Name(),
		delimiter: ":",
	}
	if err := WithOptions(opts...)(&options); err != nil {
		tb.Fatal(err)
	}
	return &PrefixHook{
		prefix:    options.prefix,
		delimiter: options.delimiter,
	}
}

// DialHook is invoked whenever a new connection is established. Since our
// implementation doesn't require any specific actions during the dialing
// process, it delegates the call to the next underlying hook in the chain.
func (h *PrefixHook) DialHook(next redis.DialHook) redis.DialHook {
	return next
}

// ProcessHook is invoked each time a new Redis command is processed. It
// transparently prefixes the key in the command under the following
// assumptions: each Redis command references its key only once and the key is
// always the second argument in the command. Commands that do not meet these
// criteria (i.e., those with fewer arguments) are left unchanged.
func (h *PrefixHook) ProcessHook(next redis.ProcessHook) redis.ProcessHook {
	return func(ctx context.Context, cmd redis.Cmder) error {
		if prev := h.prefixKey(cmd); prev != "" {
			defer h.restoreKey(cmd, prev)
		}
		if err := next(ctx, cmd); err != nil {
			return err
		}
		return nil
	}
}

// ProcessPipelineHook is invoked for pipelined Redis requests. It serves as
// the pipeline-specific counterpart to ProcessHook by prefixing each command
// within the pipeline. This ensures that all commands are properly namespaced,
// preventing data races and conflicts when tests are executed in parallel.
func (h *PrefixHook) ProcessPipelineHook(next redis.ProcessPipelineHook) redis.ProcessPipelineHook {
	return func(ctx context.Context, cmds []redis.Cmder) error {
		for _, cmd := range cmds {
			if prev := h.prefixKey(cmd); prev != "" {
				defer h.restoreKey(cmd, prev)
			}
		}
		if err := next(ctx, cmds); err != nil {
			return err
		}
		return nil
	}
}

func (h *PrefixHook) prefixKey(cmd redis.Cmder) string {
	args := cmd.Args()
	if len(args) <= 1 {
		return ""
	}
	key, ok := args[1].(string)
	if !ok {
		return ""
	}
	args[1] = h.prefix + h.delimiter + key
	return key
}

func (h *PrefixHook) restoreKey(cmd redis.Cmder, prev string) {
	args := cmd.Args()
	args[1] = prev
}

// Option configures the PrefixHook.
type Option func(*options) error

type options struct {
	prefix    string
	delimiter string
}

// WithDelimiter overrides the default delimiter ":" with a custom delimiter;
// cannot be empty.
func WithDelimiter(delimiter string) Option {
	return func(o *options) error {
		if delimiter == "" {
			return errors.New("WithDelimiter: cannot be empty")
		}
		o.delimiter = delimiter
		return nil
	}
}

// WithOptions wraps the given options into one which is useful for delegating
// options at higher levels without having to worry about append.
func WithOptions(opts ...Option) Option {
	return func(o *options) error {
		for _, opt := range opts {
			if err := opt(o); err != nil {
				// Do not decorate the returned error with any additional message as
				// nested `WithOptions` calls would create confusing error outputs.
				return err
			}
		}
		return nil
	}
}
