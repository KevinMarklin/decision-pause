package auth

import (
	"encoding/hex"
	"errors"
	"strings"
	"testing"
	"time"
)

// Фикстура посчитана независимой реализацией (.NET HMACSHA256), а не этим кодом:
//
//	secret = HMAC_SHA256("WebAppData", token)
//	hash   = hex(HMAC_SHA256(secret, "auth_date=1700000000\nquery_id=fixture-query\nuser={\"id\":777,\"first_name\":\"Test\"}"))
const (
	fixtureToken = "123456:TEST-TOKEN-FIXTURE"
	fixtureHash  = "ec8f3647f377e7219a81769875dad10746ed3757768e613763edce2c991dff4e"
	fixtureUser  = "%7B%22id%22%3A777%2C%22first_name%22%3A%22Test%22%7D"
	fixtureRaw   = "auth_date=1700000000&query_id=fixture-query&user=" + fixtureUser + "&hash=" + fixtureHash
)

// fixtureNow — момент внутри окна жизни подписи (auth_date + 60 c).
var fixtureNow = time.Unix(1_700_000_000, 0).Add(time.Minute)

func TestValidate_Fixture(t *testing.T) {
	got, err := Validate(fixtureRaw, fixtureToken, fixtureNow)
	if err != nil {
		t.Fatalf("Validate: %v", err)
	}
	if got.UserID != 777 {
		t.Errorf("UserID = %d, want 777", got.UserID)
	}
	if got.QueryID != "fixture-query" {
		t.Errorf("QueryID = %q, want fixture-query", got.QueryID)
	}
	if !got.AuthDate.Equal(time.Unix(1_700_000_000, 0)) {
		t.Errorf("AuthDate = %v", got.AuthDate)
	}
}

func TestValidate_TamperedPayload(t *testing.T) {
	// Меняем значение user, не пересчитывая hash.
	tampered := strings.Replace(fixtureRaw, "%3A%22Test%22", "%3A%22TestX%22", 1)
	if tampered == fixtureRaw {
		t.Fatal("фикстура не изменилась — тест бессмысленен")
	}
	if _, err := Validate(tampered, fixtureToken, fixtureNow); !errors.Is(err, ErrSignature) {
		t.Errorf("err = %v, want ErrSignature", err)
	}
}

func TestValidate_WrongToken(t *testing.T) {
	if _, err := Validate(fixtureRaw, "wrong-token", fixtureNow); !errors.Is(err, ErrSignature) {
		t.Errorf("err = %v, want ErrSignature", err)
	}
}

func TestValidate_DuplicateKey(t *testing.T) {
	raw := fixtureRaw + "&query_id=other"
	if _, err := Validate(raw, fixtureToken, fixtureNow); !errors.Is(err, ErrMalformed) {
		t.Errorf("err = %v, want ErrMalformed", err)
	}
}

func TestValidate_MissingHash(t *testing.T) {
	raw := "auth_date=1700000000&query_id=fixture-query&user=" + fixtureUser
	if _, err := Validate(raw, fixtureToken, fixtureNow); !errors.Is(err, ErrMalformed) {
		t.Errorf("err = %v, want ErrMalformed", err)
	}
}

func TestValidate_MalformedHash(t *testing.T) {
	raw := strings.Replace(fixtureRaw, fixtureHash, "zz"+fixtureHash[2:], 1)
	if _, err := Validate(raw, fixtureToken, fixtureNow); !errors.Is(err, ErrSignature) {
		t.Errorf("err = %v, want ErrSignature", err)
	}
}

func TestValidate_MissingUser(t *testing.T) {
	// Подпись считаем корректной отдельно, чтобы дойти до разбора user.
	raw := signFixture(t, "auth_date=1700000000\nquery_id=fixture-query")
	if _, err := Validate(raw, fixtureToken, fixtureNow); !errors.Is(err, ErrMalformed) {
		t.Errorf("err = %v, want ErrMalformed", err)
	}
}

func TestValidate_Expired(t *testing.T) {
	now := time.Unix(1_700_000_000, 0).Add(MaxAge + time.Second)
	if _, err := Validate(fixtureRaw, fixtureToken, now); !errors.Is(err, ErrExpired) {
		t.Errorf("err = %v, want ErrExpired", err)
	}
}

func TestValidate_FutureAuthDate(t *testing.T) {
	now := time.Unix(1_700_000_000, 0).Add(-10 * time.Minute)
	if _, err := Validate(fixtureRaw, fixtureToken, now); !errors.Is(err, ErrExpired) {
		t.Errorf("err = %v, want ErrExpired", err)
	}
}

func TestValidate_EmptyInputs(t *testing.T) {
	if _, err := Validate("", fixtureToken, fixtureNow); !errors.Is(err, ErrMalformed) {
		t.Errorf("raw empty: err = %v, want ErrMalformed", err)
	}
	if _, err := Validate(fixtureRaw, "", fixtureNow); !errors.Is(err, ErrMalformed) {
		t.Errorf("token empty: err = %v, want ErrMalformed", err)
	}
}

// signFixture собирает валидную строку для произвольного launch_params
// (нужна, чтобы проверить разбор полей, минуя проверку подписи).
func signFixture(t *testing.T, launchParams string) string {
	t.Helper()
	secret := hmacSHA256([]byte("WebAppData"), []byte(fixtureToken))
	hash := hex.EncodeToString(hmacSHA256(secret, []byte(launchParams)))
	return "auth_date=1700000000&query_id=fixture-query&hash=" + hash
}
