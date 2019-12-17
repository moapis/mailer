// Copyright (c) 2019, Mohlmann Solutions SRL. All rights reserved.
// Use of this source code is governed by a License that can be found in the LICENSE file.
// SPDX-License-Identifier: BSD-3-Clause

//Package mailer is a functional wrapper around the standard "net/smtp" and "html/template" packages.
package mailer

import (
	"bytes"
	"fmt"
	"html/template"
	"net/smtp"
	"strings"
)

const (
	// GlobalHeaders is set on every message.
	GlobalHeaders = "MIME-Version: 1.0\r\nContent-type: text/html; charset=\"UTF-8\"\r\n\r\n"
)

func mailHeaders(h map[string][]string) *bytes.Buffer {
	var msg bytes.Buffer
	for k, v := range h {
		if v != nil {
			msg.WriteString(fmt.Sprintf(
				"%s: %s\r\n",
				strings.Title(k),
				strings.Join(v, ","),
			))
		}
	}
	msg.WriteString(GlobalHeaders)
	return &msg
}

// Mailer holds a html template, server and authentication information for efficient reuse.
type Mailer struct {
	tmpl *template.Template
	addr string
	from string
	auth smtp.Auth
}

// New returns a reusable mailer.
// Tmpl should hold a collection of parsed templates.
// Addr is the hostname and port used by smtp.SendMail. For example:
//   "mail.host.com:587"
// From is used in every subsequent SendMail invocation.
// If auth is nil, connections will omit authentication.
func New(tmpl *template.Template, addr, from string, auth smtp.Auth) *Mailer {
	return &Mailer{tmpl, addr, from, auth}
}

// Send renders the headers and named template with passed data.
// The rendered message is sent using smtp.SendMail, to all the recipients.
//
// Headers keys are rendered Title cased, and the values are joined with a comma separator.
// Each entry becomes a CRLF separated line. For example:
//   map[string][]string{"to": []string{"foo@bar.com", "hello@world.com"}}
// Results in:
//   To: foo@bar.com,hello@world.com\r\n
func (m *Mailer) Send(headers map[string][]string, tmplName string, data interface{}, recipients []string) error {
	msg := mailHeaders(headers)
	if err := m.tmpl.ExecuteTemplate(msg, tmplName, data); err != nil {
		return err
	}
	fmt.Println(msg)
	return smtp.SendMail(m.addr, m.auth, m.from, recipients, msg.Bytes())
}
