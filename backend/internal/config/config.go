package config

import (
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	ListenAddr       string        `envconfig:"LISTEN_ADDR" default:":8080"`
	DatabaseURL      string        `envconfig:"DATABASE_URL" required:"true"`
	BaseURL          string        `envconfig:"BASE_URL" default:"http://localhost:8080"`
	CORSOrigins      string        `envconfig:"CORS_ORIGINS" default:"http://localhost:3000"`
	GracePeriod      time.Duration `envconfig:"GRACE_PERIOD" default:"2m"`
	MaxTextSize      int64         `envconfig:"MAX_TEXT_SIZE" default:"102400"`
	MaxFileSize      int64         `envconfig:"MAX_FILE_SIZE" default:"5242880"`
	CleanupInterval  time.Duration `envconfig:"CLEANUP_INTERVAL" default:"60s"`

	// S3 / MinIO
	S3Endpoint  string `envconfig:"S3_ENDPOINT" default:"localhost:9000"`
	S3AccessKey string `envconfig:"S3_ACCESS_KEY" required:"true"`
	S3SecretKey string `envconfig:"S3_SECRET_KEY" required:"true"`
	S3Bucket    string `envconfig:"S3_BUCKET" default:"sharepwd"`
	S3UseSSL    bool   `envconfig:"S3_USE_SSL" default:"false"`
	S3Region    string `envconfig:"S3_REGION" default:"us-east-1"`

	// Storage backend: "s3" or "local"
	StorageBackend string `envconfig:"STORAGE_BACKEND" default:"s3"`
	LocalStorePath string `envconfig:"LOCAL_STORE_PATH" default:"/data/files"`

	// Rate limiting
	RateLimitPublic int `envconfig:"RATE_LIMIT_PUBLIC" default:"30"`

	// Anti-bot defense layers
	ChallengeMinSolveTime time.Duration `envconfig:"CHALLENGE_MIN_SOLVE_TIME" default:"1500ms"`
	ChallengeTTL          time.Duration `envconfig:"CHALLENGE_TTL" default:"5m"`
	PowDifficulty         uint8         `envconfig:"POW_DIFFICULTY" default:"20"`
	BehavioralMinScore    int           `envconfig:"BEHAVIORAL_MIN_SCORE" default:"30"`
	EnvMinScore           int           `envconfig:"ENV_MIN_SCORE" default:"20"`
	MaxNoncesPerIP        int           `envconfig:"MAX_NONCES_PER_IP" default:"3"`
	MetadataRateLimit     int           `envconfig:"METADATA_RATE_LIMIT" default:"10"`
	DefenseStrictMode     bool          `envconfig:"DEFENSE_STRICT_MODE" default:"false"`
}

func Load() (*Config, error) {
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
