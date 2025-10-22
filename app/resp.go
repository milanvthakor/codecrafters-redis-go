package main

import (
	"fmt"
	"net"
	"strconv"
)

// EncType represents the different encoding types of the RESP value.
type EncType int

const (
	SimpleStr EncType = iota
	SimpleErr
	Integers
	BulkStrs
	Arrs
	Nulls
	Bools
	Doubles
	BigNums
	BulkErrs
)

// RespVal represents the decoded RESP value.
type RespVal struct {
	Typ EncType
	Val any
}

func (i *RespVal) SimpleStr() string {
	return i.Val.(string)
}

func (i *RespVal) SimpleErr() string {
	return i.Val.(string)
}

func (i *RespVal) Integers() int64 {
	return i.Val.(int64)
}

func (i *RespVal) BulkStrs() string {
	return i.Val.(string)
}

func (i *RespVal) ArrElems() []*RespVal {
	return i.Val.([]*RespVal)
}

func (i *RespVal) Bools() bool {
	return i.Val.(bool)
}

func (i *RespVal) Doubles() float64 {
	return i.Val.(float64)
}

func (i *RespVal) BigNums() string {
	return i.Val.(string)
}

func (i *RespVal) BulkErrs() string {
	return i.Val.(string)
}

// readRespVal reads the input command from the connection as per the RESP format
func readRespVal(conn net.Conn) (*RespVal, error) {
	c := &RespVal{}

	input, err := readUntilCRLF(conn)
	if err != nil {
		return nil, err
	} else if len(input) <= 0 {
		return c, nil
	}

	switch input[0] {
	case '+':
		c.Typ = SimpleStr
		c.Val = input[1:]

	case '-':
		c.Typ = SimpleErr
		c.Val = input[1:]

	case ':':
		c.Typ = Integers
		intVal, err := strconv.ParseInt(input[1:], 10, 64)
		if err != nil {
			return nil, err
		}

		c.Val = intVal

	case '$':
		c.Typ = BulkStrs
		str, err := readUntilCRLF(conn)
		if err != nil {
			return nil, err
		}

		c.Val = str

	case '*':
		c.Typ = Arrs
		arrSize, err := strconv.Atoi(input[1:])
		if err != nil {
			return nil, err
		}

		// Read the array elements
		arrElems := make([]*RespVal, 0, arrSize)
		for range arrSize {
			elem, err := readRespVal(conn)
			if err != nil {
				return nil, err
			}

			arrElems = append(arrElems, elem)
		}

		c.Val = arrElems

	case '_':
		c.Typ = Nulls

	case '#':
		c.Typ = Bools
		if input[1:] == "t" {
			c.Val = true
		} else {
			c.Val = false
		}

	case ',':
		c.Typ = Doubles
		floatVal, err := strconv.ParseFloat(input[1:], 64)
		if err != nil {
			return nil, err
		}

		c.Val = floatVal

	case '(':
		c.Typ = BigNums
		c.Val = input[1:] // values could range outside of 64 bits. Hence, storing it as string

	case '!':
		c.Typ = BulkErrs
		c.Val = input[1:]

	default:
		return nil, fmt.Errorf("invalid command")
	}

	return c, nil
}

func ToSimpleStr(val string) string {
	return fmt.Sprintf("+%s\r\n", val)
}

func ToBulkStr(val any) string {
	return fmt.Sprintf("$%d\r\n%v\r\n", len(fmt.Sprint(val)), val)
}

func ToNulls() string {
	return "$-1\r\n"
}

func ToIntegers(val int) string {
	return fmt.Sprintf(":%d\r\n", val)
}

func ToArray(arr []any) string {
	if arr == nil {
		return "*-1\r\n"
	}

	str := fmt.Sprintf("*%d\r\n", len(arr))
	for _, ele := range arr {
		str += fmt.Sprintf("$%d\r\n%v\r\n", len(fmt.Sprint(ele)), ele)
	}

	return str
}
