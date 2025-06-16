package mail

import (
	"os"
	"strconv"

	"gopkg.in/gomail.v2"
)


type MailService struct {
	dialer *gomail.Dialer
}

func NewMailService() *MailService {
	host := os.Getenv("SMTP_HOST")
	port, _ := strconv.Atoi(os.Getenv("SMTP_PORT"))
	user := os.Getenv("SMTP_USER")
	pass := os.Getenv("SMTP_PASSWORD")
	return &MailService{
		dialer: gomail.NewDialer(host, port, user, pass),
	}
}

func (m *MailService) SendActivationMail(to, activationLink string) error {
	message := gomail.NewMessage()
	message.SetHeader("From", os.Getenv("SMTP_USER"))
	message.SetHeader("To", to)
	message.SetHeader("Subject", "Активация аккаунта в Lumivy")

	body := `
<!DOCTYPE html>
<html lang="ru">
<head>
<meta charset="UTF-8">
<title>Активация Lumivy</title>
</head>
<body style="margin:0;padding:0;background:#f3e5f5;font-family:Arial,sans-serif;">
  <table role="presentation" cellpadding="0" cellspacing="0" width="100%">
	<tr>
	  <td align="center" style="padding:40px 15px;">
		<table role="presentation" cellpadding="0" cellspacing="0" width="600" style="max-width:600px;background:#ffffff;border-radius:10px;overflow:hidden;box-shadow:0 4px 16px rgba(74,0,128,.12);">
		  <tr>
			<td style="height:6px;background:linear-gradient(90deg,#ba68c8 0%,#8e24aa 100%);"></td>
		  </tr>
		  <tr>
			<td style="padding:32px 40px 24px 40px;text-align:center;">
			  <h2 style="margin:0 0 12px 0;color:#8e24aa;font-size:24px;font-weight:700;">
				Добро пожаловать в Lumivy
			  </h2>
			  <p style="margin:0 0 24px 0;color:#503e5d;font-size:16px;line-height:1.5;">
				Спасибо за регистрацию! Чтобы активировать аккаунт, нажмите кнопку ниже.
			  </p>
			  <a href="` + activationLink + `" style="display:inline-block;padding:14px 32px;background:#8e24aa;color:#ffffff;text-decoration:none;border-radius:40px;font-weight:600;">
				Активировать аккаунт
			  </a>
			  <p style="margin:32px 0 0 0;color:#7b6a85;font-size:13px;line-height:1.4;">
				Если кнопка не работает, скопируйте ссылку в адресную строку браузера:<br>
				<span style="word-break:break-all;color:#8e24aa;">` + activationLink + `</span>
			  </p>
			</td>
		  </tr>
		</table>
	  </td>
	</tr>
  </table>
</body>
</html>`
	message.SetBody("text/html", body)
	return m.dialer.DialAndSend(message)
}

func (m *MailService) SendFamilyInviteMail(to, inviteLink string) error {
	message := gomail.NewMessage()
	message.SetHeader("From", os.Getenv("SMTP_USER"))
	message.SetHeader("To", to)
	message.SetHeader("Subject", "Приглашение в семью Lumivy")

	body := `
<!DOCTYPE html>
<html lang="ru">
<head>
<meta charset="UTF-8">
<title>Приглашение Lumivy</title>
</head>
<body style="margin:0;padding:0;background:#f3e5f5;font-family:Arial,sans-serif;">
  <table role="presentation" cellpadding="0" cellspacing="0" width="100%">
	<tr>
	  <td align="center" style="padding:40px 15px;">
		<table role="presentation" cellpadding="0" cellspacing="0" width="600" style="max-width:600px;background:#ffffff;border-radius:10px;overflow:hidden;box-shadow:0 4px 16px rgba(74,0,128,.12);">
		  <tr>
			<td style="height:6px;background:linear-gradient(90deg,#ba68c8 0%,#8e24aa 100%);"></td>
		  </tr>
		  <tr>
			<td style="padding:32px 40px 24px 40px;text-align:center;">
			  <h2 style="margin:0 0 12px 0;color:#8e24aa;font-size:24px;font-weight:700;">
				Вас пригласили присоединиться к семье
			  </h2>
			  <p style="margin:0 0 24px 0;color:#503e5d;font-size:16px;line-height:1.5;">
				Для подтверждения нажмите кнопку ниже. Если у вас ещё нет аккаунта — зарегистрируйтесь и вернитесь по ссылке.
			  </p>
			  <a href="` + inviteLink + `" style="display:inline-block;padding:14px 32px;background:#8e24aa;color:#ffffff;text-decoration:none;border-radius:40px;font-weight:600;">
				Принять приглашение
			  </a>
			  <p style="margin:32px 0 0 0;color:#7b6a85;font-size:13px;line-height:1.4;">
				Если кнопка не работает, скопируйте ссылку в браузер:<br>
				<span style="word-break:break-all;color:#8e24aa;">` + inviteLink + `</span>
			  </p>
			</td>
		  </tr>
		</table>
	  </td>
	</tr>
  </table>
</body>
</html>`
	message.SetBody("text/html", body)
	return m.dialer.DialAndSend(message)
}
