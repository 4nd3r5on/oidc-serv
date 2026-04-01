package main

import "os"

func getEnv(name, fallback string) string {
	val, ok := os.LookupEnv(name)
	if !ok {
		return fallback
	}
	return val
}
