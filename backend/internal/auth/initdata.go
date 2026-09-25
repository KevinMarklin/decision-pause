// Package auth проверяет подписанные стартовые параметры мини-аппа MAX
// (window.WebApp.initData) и извлекает из них id пользователя.
//
// Алгоритм — по спеке https://dev.max.ru/docs/webapps/validation:
//
//	secret      = HMAC-SHA256(key="WebAppData", data=botToken)
//	hash        = hex(HMAC-SHA256(key=secret, data=launch_params))
//
// launch_params — пары key=value без hash, с URL-декодированными значениями,
// отсортированные по ключу и склеенные через "\n".
package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

// MaxAge — максимальный возраст подписи. MAX рекомендует 1 час.
const MaxAge = time.Hour

// MaxFutureSkew — допустимое «опережение» auth_date относительно часов сервера.
const MaxFutureSkew = 5 * time.Minute

// Ошибки валидации. Содержат только категорию, никогда — сырые данные или токен.
var (
	ErrMalformed = errors.New("init data malformed")
	ErrSignature = errors.New("init data signature mismatch")
	ErrExpired   = errors.New("init data expired")
)

// InitData — проверенные стартовые параметры мини-аппа.
type InitData struct {
	UserID     int64
	AuthDate   time.Time
	QueryID    string
	StartParam string
}

// Validate проверяет подпись raw токеном botToken и возвращает данные.
// now передаётся явно, чтобы тесты были детерминированы.
func Validate(raw, botToken string, now time.Time) (InitData, error) {
	if raw == "" || botToken == "" {
		return InitData{}, ErrMalformed
	}

	// 1–2. Разбор пар key=value, значения URL-декодируем (один проход,
	// как decodeURIComponent в примере из доков).
	pairs := make(map[string]string, 8)
	for _, part := range strings.Split(raw, "&") {
		if part == "" {
			continue
		}
		key, value, ok := strings.Cut(part, "=")
		if !ok || key == "" {
			return InitData{}, ErrMalformed
		}
		if _, dup := pairs[key]; dup {
			return InitData{}, ErrMalformed // дубликат ключа — атака на разбор
		}
		decoded, err := url.PathUnescape(value)
		if err != nil {
			return InitData{}, ErrMalformed
		}
		pairs[key] = decoded
	}

	// 3. Ровно один hash, он же не участвует в подписываемой строке.
	originalHash, ok := pairs["hash"]
	if !ok {
		return InitData{}, ErrMalformed
	}
	delete(pairs, "hash")

	// 5–6. Сортировка по ключу и склейка launch_params.
	keys := make([]string, 0, len(pairs))
	for k := range pairs {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var b strings.Builder
	for i, k := range keys {
		if i > 0 {
			b.WriteByte('\n')
		}
		b.WriteString(k)
		b.WriteByte('=')
		b.WriteString(pairs[k])
	}
	launchParams := b.String()

	// 7–9. Двухшаговая деривация ключа — как в Telegram Mini Apps.
	secret := hmacSHA256([]byte("WebAppData"), []byte(botToken))
	computed := hmacSHA256(secret, []byte(launchParams))

	want, err := hex.DecodeString(originalHash)
	if err != nil || len(want) != sha256.Size {
		return InitData{}, ErrSignature
	}
	if !hmac.Equal(computed, want) {
		return InitData{}, ErrSignature
	}

	// Дальше данные считаем подлинными и разбираем payload.
	out := InitData{QueryID: pairs["query_id"], StartParam: pairs["start_param"]}

	authDateRaw, ok := pairs["auth_date"]
	if !ok {
		return InitData{}, ErrMalformed
	}
	sec, err := strconv.ParseInt(authDateRaw, 10, 64)
	if err != nil {
		return InitData{}, ErrMalformed
	}
	authDate := time.Unix(sec, 0)
	if now.After(authDate.Add(MaxAge)) {
		return InitData{}, ErrExpired
	}
	if authDate.After(now.Add(MaxFutureSkew)) {
		return InitData{}, ErrExpired
	}
	out.AuthDate = authDate

	userRaw, ok := pairs["user"]
	if !ok {
		return InitData{}, ErrMalformed
	}
	var user struct {
		ID int64 `json:"id"`
	}
	if err := json.Unmarshal([]byte(userRaw), &user); err != nil {
		return InitData{}, ErrMalformed
	}
	if user.ID <= 0 {
		return InitData{}, ErrMalformed
	}
	out.UserID = user.ID

	return out, nil
}

func hmacSHA256(key, data []byte) []byte {
	mac := hmac.New(sha256.New, key)
	mac.Write(data)
	return mac.Sum(nil)
}
