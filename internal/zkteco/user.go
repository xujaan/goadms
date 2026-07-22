package zkteco

import (
	"fmt"
	"strings"
)

// UserRecord represents a user stored on a device.
type UserRecord struct {
	UserID    string `json:"user_id"`
	Name      string `json:"name"`
	CardNo    string `json:"card_no"`
	Password  string `json:"password"`
	Privilege int    `json:"privilege"`
}

// GetUsers fetches all users from the device.
func (c *Client) GetUsers() ([]UserRecord, error) {
	data, err := c.ExecData(CmdGetUser, nil)
	if err != nil {
		return nil, err
	}
	return parseUserData(data)
}

// SetUser adds or updates a user on the device.
func (c *Client) SetUser(u UserRecord) error {
	// Format: "user_id\tname\tpassword\tprivilege\tcard_no"
	payload := fmt.Sprintf("%s\t%s\t%s\t%d\t%s",
		u.UserID, u.Name, u.Password, u.Privilege, u.CardNo,
	)
	return c.ExecSimple(CmdSetUser, []byte(payload))
}

// DeleteUser removes a user from the device by user ID.
func (c *Client) DeleteUser(userID string) error {
	return c.ExecSimple(CmdDelUser, []byte(userID))
}

func parseUserData(data []byte) ([]UserRecord, error) {
	if len(data) == 0 {
		return nil, nil
	}

	lines := strings.Split(string(data), "\n")
	var users []UserRecord

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		fields := strings.Split(line, "\t")
		if len(fields) < 2 {
			continue
		}

		u := UserRecord{
			UserID: strings.TrimSpace(fields[0]),
			Name:   strings.TrimSpace(fields[1]),
		}
		if len(fields) > 2 {
			u.Password = strings.TrimSpace(fields[2])
		}
		if len(fields) > 3 {
			u.Privilege = atoi(strings.TrimSpace(fields[3]))
		}
		if len(fields) > 4 {
			u.CardNo = strings.TrimSpace(fields[4])
		}

		users = append(users, u)
	}

	return users, nil
}
