package utils

/*
Provides MsgIDGenerator to generate exim-compliant message id.
Based on "4. Message identification"
@ http://www.exim.org/exim-html-current/doc/html/spec_html/ch-how_exim_receives_and_delivers_mail.html

Applies base62 encoding with the following parts (separated by hyphens):
	- 6 chars for seconds since epoch.
	- 6 chars for process id.
	- 2 chars for nanoseconds modulus max number.
*/

import (
	"fmt"
	"os"
	"strings"
	"time"

	"bitbucket.org/fusemail/fm-lib-commons-golang/utils/basecoder"
	log "github.com/sirupsen/logrus"
)

// MsgIDGenerator provides Generate() to generate message id based on fields.
type MsgIDGenerator struct {
	ProcessID int
	Date      time.Time

	// Backup date (now by default) to generate nanoseconds as the last msgid part,
	// only in case Date does NOT have nanoseconds.
	BackupDate time.Time
}

// NewMsgIDGenerator constructs MsgIDGenerator instance with defaults.
func NewMsgIDGenerator(date time.Time) *MsgIDGenerator {
	return &MsgIDGenerator{
		Date:       date,
		BackupDate: time.Now(),
		ProcessID:  os.Getpid(),
	}
}

func (m *MsgIDGenerator) String() string {
	return fmt.Sprintf("{%T: %d on %v (backup: %v)}", m, m.ProcessID, m.Date, m.BackupDate)
}

// Generate generates message id based on struct field values.
func (m *MsgIDGenerator) Generate() string {
	logger := log.WithField("generator", m)

	// Date could have nanos truncated, in which case assign backup's.
	nanos := m.Date.Nanosecond()
	if nanos == 0 {
		nanos = m.BackupDate.Nanosecond() // Apply backup.

		// Apply now's nanos as last resort.
		if nanos == 0 {
			nanos = time.Now().Nanosecond()
		}
	}

	coder62 := basecoder.New62()
	base := 62

	parts := []int{
		int(m.Date.Unix()),
		m.ProcessID,
		nanos % (base * base),
	}

	logger = logger.WithField("parts", parts)
	logger.Debug("parts")

	msgid := fmt.Sprintf("%06s-%06s-%02s",
		coder62.Encode(parts[0]),
		coder62.Encode(parts[1]),
		coder62.Encode(parts[2]),
	)

	logger = logger.WithField("msgid", msgid)
	logger.Debug("msgid")

	return msgid
}

// GenerateMsgID generates exim-compliant message id based on provided date and other defaults.
func GenerateMsgID(date time.Time) string {
	return NewMsgIDGenerator(date).Generate()
}

// IsEximMsgID returns true if input msgid contains 6-6-2 base62 characters
// (with hyphens); else it returns false (including empty string).
// NOTE: function only checks first 16 bytes of input msgid, so suffix a valid
//       msgid does not effect the return result of this function.
func IsEximMsgID(msgid string) bool {
	if len(msgid) < 16 {
		return false
	}
	tokens := strings.Split(msgid, "-")
	if len(tokens) < 3 {
		return false
	}
	if len(tokens[0]) != 6 || !basecoder.IsBase62(tokens[0]) {
		return false
	}
	if len(tokens[1]) != 6 || !basecoder.IsBase62(tokens[1]) {
		return false
	}

	return (len(tokens[2]) >= 2 && basecoder.IsBase62(tokens[2][:2]))
}
