package auth

import (
	"fmt"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	jwt "github.com/golang-jwt/jwt/v4"
)

func AuthenticateToken() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("authorization")
		if authHeader == "" {
			send401Res(c, "couldn't authenticate: authorization header isn't populated")
			return
		}
		tokenString := strings.Split(authHeader, " ")[1]

		//parse our token. populates the sig and sets a boolean indicating its valid
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			//use a type assertion to make sure the method we used to sign this token is what we expect
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("error parsing the token")
			}
			return []byte(os.Getenv("ACCESS_TOKEN_SECRET")), nil
		})

		if token == nil || err != nil {
			fmt.Println("err", err)
			send401Res(c, "invalid access token")
			return
		}

		//use a type assertion to make sure the claims are in the format we expect
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			send401Res(c, "invalid token: couldn't parse claims")
			return
		}
		username := claims["username"].(string)
		if username != "agent" {
			send401Res(c, "invalid token: couldn't parse claims")
			return
		}
		c.Next()
	}

}

func send401Res(c *gin.Context, msg string) {
	c.AbortWithStatusJSON(401, msg)
}
