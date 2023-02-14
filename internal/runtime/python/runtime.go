package python

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
)

const (
	runtimeName = "python"

	runtimeWrapper = `
from main import handler
import sys, json

class Logger:
	def log(self, message):
		print(message, file=sys.stderr)	

class Context(dict):
    def __init__(self, *args, **kwargs):
        super(Context, self).__init__(*args, **kwargs)
        self.logger = Logger()

    def __getattr__(self, name):
        try:
            return self[name]
        except KeyError:
            raise AttributeError(f"'{self.__class__.__name__}' object has no attribute '{name}'")

ctx = Context()

params = json.loads(sys.argv[1]);

print(json.dumps(handler(ctx, params)))
`
)

type runtime struct {
	Logger *log.Entry
}

func New() (*runtime, error) {
	r := &runtime{}

	// Check for runtime version on the host.
	// If an error occurs here, it potentially means that
	// the underlying tool isn't installed or can't be found.
	version, err := r.Version()
	if err != nil {
		return nil, err
	}

	r.Logger = log.New().
		WithField("runtime", r.Name()).
		WithField("version", version)

	r.Logger.Info("Runtime initialized")

	return r, nil
}

// Name return the name of the current runtime.
func (r *runtime) Name() string {
	return runtimeName
}

// Version retrieve the version of the current runtime on host.
// An error can be returned if the executable can't be found in $PATH,
// or if the command can't be executed for any reasons.
func (r *runtime) Version() (string, error) {
	cmd := exec.Command("python", "-c", "import sys; print(sys.version[:6])")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// WrapCmd will set up all the required stuff inside the working directory, like
// installing dependencies, injecting wrapper etc.
func (r *runtime) WrapCmd(ctx context.Context) (*exec.Cmd, error) {
	wd := ctx.Value("wd").(string)
	iid := ctx.Value("iid").(string)

	logger := r.Logger.WithField("wd", wd).WithField("iid", iid)

	// First, we need to check for a requirements.txt file inside the current working directory
	// and if it exists, we run the dependencies installation task
	if _, err := os.Stat(filepath.Join(wd, "requirements.txt")); !os.IsNotExist(err) {
		logger.Debug("requirements.txt file detected inside the working directory")
		if err := r.installDependencies(wd); err != nil {
			return nil, err
		}
	}

	// Inject the runtime wrapper inside the working directory.
	// Currently, we assume that a main.py file is present into the working directory.
	// The main.py file include a function called `handler` in order to be executed.
	// We need to inject a custom wrapper in order to pass context / variables to our function.
	trigger := fmt.Sprintf("%s.py", iid)
	if err := os.WriteFile(filepath.Join(wd, trigger), []byte(runtimeWrapper), 0644); err != nil {
		panic(err)
	}

	logger.Debug("Wrapper injected into function working directory")

	return exec.Command("python", trigger), nil
}

func (r *runtime) installDependencies(wd string) error {
	r.Logger.Debug("Installing dependencies")
	cmd := exec.Command("pip", "install", "-r", "requirements.txt")
	cmd.Dir = wd
	out, err := cmd.CombinedOutput()
	r.Logger.Trace(string(out))
	r.Logger.Debug("Dependencies installed")
	return err
}
