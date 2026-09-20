package hashing

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
)

func GetHashedBody(key string, body []byte) ([]byte, error) {
	h := hmac.New(sha256.New, []byte(key))
	h.Write(body)
	s := h.Sum(nil)
	return s, nil
}

func CheckHashedBody(key string, hashedData string, body []byte) (bool, error) {
	dataRequest, err := GetHashedBody(key, body)
	if err != nil {
		return false, fmt.Errorf("error get hash body: %w", err)
	}
	msg, err := base64.StdEncoding.DecodeString(hashedData)
	if err != nil {
		return false, fmt.Errorf("error decode hash from header: %w", err)
	}

	if hmac.Equal(msg, dataRequest) {
		return true, nil
	} else {
		return false, nil
	}
}

func HashHandler(key string, h http.HandlerFunc) http.HandlerFunc {
	if key == "" {
		return h
	} else {
		fn := func(w http.ResponseWriter, r *http.Request) {
			contentHash := r.Header.Get("HashSHA256")
			if contentHash != "" {
				body, err := io.ReadAll(r.Body)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				signValide, err := CheckHashedBody(key, contentHash, body)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				if !signValide {
					w.WriteHeader(http.StatusBadRequest)
					return
				}

				r.Body = io.NopCloser(bytes.NewBuffer(body))
			}

			hashResponseWriter := &HashResponseWriter{
				w: w,
				b: &bytes.Buffer{},
			}

			h.ServeHTTP(hashResponseWriter, r)

			w = hashResponseWriter.w

			b, err := GetHashedBody(key, hashResponseWriter.b.Bytes())
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			w.Header().Set("HashSHA256", base64.StdEncoding.EncodeToString(b))

			_, err = w.Write(hashResponseWriter.b.Bytes())
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}

		return fn
	}
}
