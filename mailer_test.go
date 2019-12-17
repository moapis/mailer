package mailer

import (
	"html/template"
	"net/smtp"
	"reflect"
	"testing"
)

const testMailHeadersOut = "To: foo@bar.com,hello@world.com\r\nFrom: info@spanac.ro\r\n" + GlobalHeaders

func Test_mailHeaders(t *testing.T) {
	tests := []struct {
		name string
		h    map[string][]string
		want []byte
	}{
		{
			"Empty",
			nil,
			[]byte(GlobalHeaders),
		},
		{
			"Mixed entries",
			map[string][]string{
				"to":      {"foo@bar.com", "hello@world.com"},
				"from":    {"info@spanac.ro"},
				"subject": nil,
			},
			[]byte(testMailHeadersOut),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := mailHeaders(tt.h).Bytes(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("New() = \n%swant\n%s", got, tt.want)
			}
		})
	}
}

const testTemplate = `{{ define "tester" }}
<html>
<body>
<h1>Hello, World!</h1>
<p>
	This is the unit tester from <a href="{{ .URL }}">moapis/mailer</a>.
	A functional wrapper around the standard Go "net/smtp" and "html/template" packages.
</p>
</body>
</html>
{{ end }}`

var testTmplData = struct{ URL string }{"https://github.com/moapis/mailer"}

func TestNew(t *testing.T) {
	type args struct {
		tmpl *template.Template
		addr string
		from string
		auth smtp.Auth
	}
	tests := []struct {
		name string
		args args
		want *Mailer
	}{
		{
			"TestNew",
			args{
				tmpl: template.Must(template.New("test").Parse(testTemplate)),
				addr: "smtp.example.com:578",
				from: "test@example.com",
				auth: nil,
			},
			&Mailer{
				tmpl: template.Must(template.New("test").Parse(testTemplate)),
				addr: "smtp.example.com:578",
				from: "test@example.com",
				auth: nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := New(tt.args.tmpl, tt.args.addr, tt.args.from, tt.args.auth); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("New() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMailer_Send(t *testing.T) {
	type fields struct {
		tmpl *template.Template
		addr string
		from string
		auth smtp.Auth
	}
	type args struct {
		headers    map[string][]string
		tmplName   string
		data       interface{}
		recipients []string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			"Template error",
			fields{
				tmpl: template.Must(template.New("test").Parse(testTemplate)),
				addr: "test.mailu.io:578",
				from: "admin@test.mailu.io",
				auth: nil,
			},
			args{
				headers:    nil,
				tmplName:   "spanac",
				data:       testTmplData,
				recipients: []string{"test@test.mailu.io", "admin@test.mailu.io"},
			},
			true,
		},
		{
			"Send error",
			fields{
				tmpl: template.Must(template.New("test").Parse(testTemplate)),
				addr: "test.mailu.io:578",
				from: "admin@test.mailu.io",
				auth: nil,
			},
			args{
				headers:    nil,
				tmplName:   "tester",
				data:       testTmplData,
				recipients: []string{"test@test.mailu.io", "admin@test.mailu.io"},
			},
			true,
		},
		{
			"Send success",
			fields{
				tmpl: template.Must(template.New("test").Parse(testTemplate)),
				addr: "test.mailu.io:587",
				from: "admin@test.mailu.io",
				auth: smtp.PlainAuth("", "admin@test.mailu.io", "letmein", "test.mailu.io"),
			},
			args{
				headers: map[string][]string{
					"to":      {"test@test.mailu.io", "admin@test.mailu.io"},
					"from":    {"admin@test.mailu.io"},
					"subject": {"moapis/mailer: Unit tests"},
				},
				tmplName:   "tester",
				data:       testTmplData,
				recipients: []string{"test@test.mailu.io", "admin@test.mailu.io"},
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Mailer{
				tmpl: tt.fields.tmpl,
				addr: tt.fields.addr,
				from: tt.fields.from,
				auth: tt.fields.auth,
			}
			if err := m.Send(tt.args.headers, tt.args.tmplName, tt.args.data, tt.args.recipients); (err != nil) != tt.wantErr {
				t.Errorf("Mailer.Send() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
