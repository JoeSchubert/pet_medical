package db

import "time"

const connectRetries = 10
const connectRetryDelay = 2 * time.Second
