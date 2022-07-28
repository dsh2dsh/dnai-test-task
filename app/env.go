package app

import (
	"github.com/joho/godotenv"
)

// LoadDotEnv reads .env files and loads their content into env variables. It
// uses env for building name of .env files. By default its value is
// "development" and in this case it loads:
//
//   1. .env.development.local
//   2. .env.local
//   3. .env.development
//   4. .env
//
// For "test" environment it skips loading of .env.local file.
func LoadDotEnv(env string) {
	if env == "" {
		env = "development"
	}
	godotenv.Load(".env." + env + ".local")
	if env != "test" {
		godotenv.Load(".env.local")
	}
	godotenv.Load(".env." + env)
	godotenv.Load() // load .env
}
