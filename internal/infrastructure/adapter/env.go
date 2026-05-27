package adapter

import (
	e "errors"
	"io/fs"
	"os"
	"time"

	"fms-project/internal/infrastructure/errors"

	"github.com/joho/godotenv"
)

func InitDotEnv() error {
	if err := godotenv.Load(); err != nil {
		if e.Is(err, fs.ErrNotExist) {
			return nil
		}
		return errors.Wrap(err, "dotenv-init", "could not load .env")
	}
	return nil
}

type resolver struct {
	key string
	def *string
}

func Getenv(key string) resolver {
	return resolver{key: key}
}

func (r resolver) Default(v string) resolver {
	r.def = &v
	return r
}

func (r resolver) ToString() string {
	if v := os.Getenv(r.key); v != "" {
		return v
	}
	if r.def != nil {
		return *r.def
	}

	panic(errors.New("env-resolver", "env not set: "+r.key))
}

func (r resolver) ToDuration() time.Duration {
	if v := os.Getenv(r.key); v != "" {
		if duration, err := time.ParseDuration(v); err == nil {
			return duration
		}
	}
	if r.def != nil {
		if duration, err := time.ParseDuration(*r.def); err == nil {
			return duration
		}
	}

	panic(errors.New("env-resolver", "env not set: "+r.key))
}

func (r resolver) ToBool() bool {
	if v := os.Getenv(r.key); v != "" {
		switch v {
		case "true":
			return true
		case "false":
			return false
		default:
			panic(errors.New("env-resolver", "env not bool: "+r.key))
		}
	}
	if r.def != nil {
		switch *r.def {
		case "true":
			return true
		case "false":
			return false
		default:
			panic(errors.New("env-resolver", "env not bool: "+r.key))
		}
	}

	panic(errors.New("env-resolver", "env not set: "+r.key))
}
