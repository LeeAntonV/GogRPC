package email

import (
	"fmt"
	"github.com/joho/godotenv"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
	"math/rand"
	"net/mail"
	"net/smtp"
	"os"
)

func SendEmail(email string) (int, error) {
	randomInt := rand.Intn(900000) + 100000
	err := godotenv.Load()
	if err != nil {
		//return 0, err
		log.Fatal(err)
	}

	////var logger *slog.Logger
	////log := logger.With(
	////	slog.String("email", email),
	////)
	//
	//log.Info("Attempting to send email")

	senderEmail := os.Getenv("SenderEmail")
	password := os.Getenv("EmailSecret")

	smtpPort := os.Getenv("EmailPort")

	smtpServer := "smtp.gmail.com"

	from := mail.Address{Address: senderEmail}
	to := mail.Address{Address: email}
	subject := "GRPC server register message"

	message := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\n\r\n%d",
		from.String(), to.String(), subject, randomInt)

	auth := smtp.PlainAuth("", senderEmail, password, smtpServer)
	err = smtp.SendMail(fmt.Sprintf("%s:%s", smtpServer, smtpPort), auth, from.Address, []string{to.Address}, []byte(message))

	if err != nil {
		//log.Error("Failed to send email", sl.Err(err))
		return 0, status.Error(codes.Internal, "Failed to send code")
	}

	return randomInt, nil
}
