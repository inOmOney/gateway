package public

import (
	"fmt"
	"testing"
)

func TestJwtDecode(t *testing.T) {
	token := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2MTcyOTI5OTUsImlzcyI6ImFwcF9pZF9hIn0.r3uQc6ZxKnt4DjYd87Fvi8VTWMo-d_c4iwqaAS93g3M"
	decode, err := JwtDecode(token)
	if err!=nil {
		panic(err)
	}
	fmt.Println(decode.Issuer)

}
