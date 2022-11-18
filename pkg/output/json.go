package output

import (
	"context"
	"fmt"
	"os"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

type jsonManager struct{}

func (j *jsonManager) Output(ctx context.Context, out interface{}) error {
	if m, ok := out.(proto.Message); ok {
		outBytes, err := protojson.Marshal(m)
		if err != nil {
			return err
		}

		_, err = fmt.Fprint(os.Stdout, string(outBytes))
		if err != nil {
			return err
		}

		return nil
	}

	return fmt.Errorf("unexpected output type")
}
