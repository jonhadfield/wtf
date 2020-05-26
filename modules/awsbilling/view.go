package awsbilling

import "fmt"

func parseAccountSettings(settings interface{}) (acc account, err error) {
	v, ok := settings.(map[string]interface{})
	if !ok {
		return acc, fmt.Errorf("invalid account format")
	}

	var alias, accessKeyId, accessKeySecret string

	var tInterface interface{}

	var budget float64

	var keyCheck, typeCheck bool

	tInterface, keyCheck = v["alias"]

	alias, typeCheck = tInterface.(string)
	if !keyCheck || !typeCheck {
		return acc, fmt.Errorf("alias is missing or invalid")
	}

	tInterface, keyCheck = v["accessKeyId"]

	accessKeyId, typeCheck = tInterface.(string)

	if !keyCheck || !typeCheck {
		return acc, fmt.Errorf("accessKeyId is missing or invalid")
	}

	tInterface, keyCheck = v["accessKeySecret"]
	accessKeySecret, typeCheck = tInterface.(string)

	if !keyCheck || !typeCheck {
		return acc, fmt.Errorf("accessKeySecret is missing or invalid")
	}

	tInterface, keyCheck = v["budget"]
	budget, typeCheck = tInterface.(float64)

	if !keyCheck || !typeCheck {
		return acc, fmt.Errorf("budget is missing or invalid")
	}

	acc = account{
		alias:           alias,
		accessKeyId:     accessKeyId,
		accessKeySecret: accessKeySecret,
		budget:          budget,
	}

	return acc, err
}

func strPtr(s string) *string {
	return &s
}
