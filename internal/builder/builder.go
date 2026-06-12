package builder

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
)

func Build(steps [][]string) error {
	for i, step := range steps {
		slog.Info(fmt.Sprintf("running build step %d", i))

		cmd := exec.Command(step[0], step[1:]...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err != nil {
			return fmt.Errorf("failed to execute build step %d: %w", i, err)
		}
	}

	return nil
}
