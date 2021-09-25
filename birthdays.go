package birthdays

import (
	"fmt"
	"io/ioutil"
	"regexp"
	"runtime/debug"
	"strconv"
	"strings"
	"time"
)

type birthdayDate struct {
	raw string
	parsed time.Time
	hasYear bool
	hasMonth bool
	hasDay bool
}

type birthdayData struct {
	raw string
	name string
	date birthdayDate
	error bool
}

// Show error message
func ShowError(a ...interface{}){
	messageBoxPlain("ERROR", fmt.Sprintln(a...) + string(debug.Stack()))
}

// Calculate next birthday date after or exactly at provided time
func (bd *birthdayDate) nextAfter(date time.Time) (res time.Time) {
	y, m, d := date.Date()
	var next_bd_year int
	if bd.parsed.Month() < m || (bd.parsed.Month() == m && bd.parsed.Day() < d) {
		next_bd_year = y + 1
	} else {
		next_bd_year = y
	}
	return time.Date(next_bd_year, bd.parsed.Month(), bd.parsed.Day(), 0, 0, 0, 0, time.Local)
}

// Check if birthday date has day and month
func (bd *birthdayDate) exact() bool {
	return bd.hasDay && bd.hasMonth
}

// Parse string "dd.mm.yyyy name" as birthday data
func ParseBirthdayData(raw string) (res birthdayData) {
	defer func(){
		if r := recover(); r != nil {
			res.error = true
		}
	}()
	res.raw = raw
	res.error = false
	splitted := strings.SplitN(raw, " ", 2)
	if len(splitted) != 2 {
		panic("Incorrect entry format.")
	}
	res.date.raw = splitted[0]
	res.name = splitted[1]
	splittedDate := strings.Split(splitted[0], ".")
	if len(splittedDate) != 3 {
		panic("Incorrect date format.")
	}
	dateComponents := [3]int{1, 1, 1}
	dateComponentPresent := [3]*bool{&res.date.hasDay, &res.date.hasMonth, &res.date.hasYear}
	for i, v := range splittedDate {
		if empty, err := regexp.MatchString(`^[?]+|[x]+$`, v); err != nil || empty {
			continue
		}
		var err error
		if dateComponents[i], err = strconv.Atoi(v); err != nil {
			panic("Incorrect date format.")
		}
		*dateComponentPresent[i] = true
	}
	res.date.parsed = time.Date(dateComponents[2], time.Month(dateComponents[1]), dateComponents[0], 0, 0, 0, 0, time.Local)
	return
}

// Standard string for birthday data
func (data *birthdayData) String() string {
	if data.error {
		return data.raw
	}
	var age string
	if data.date.hasYear {
		age = strconv.Itoa(time.Now().Year() - data.date.parsed.Year())
	} else {
		age = "?"
	}
	return fmt.Sprintf("%s - %s(%s)", data.name, data.date.raw, age)
}

// Load birthdays data from file
func load() (res []birthdayData) {
	bytes, err := ioutil.ReadFile(`BD.txt`)
	if err != nil {
		panic(err)
	}
	raw := strings.Split(string(bytes), "\n")
	res = make([]birthdayData, 0, len(raw))
	for _, v := range raw {
		skip, err := regexp.MatchString(`^((\s*)|(//.*))$`, v)
		if err == nil && skip { continue }
		res = append(res, ParseBirthdayData(v))
	}
	return
}

// Filter actual/upcoming/uncertain/erroneous birthday entries
func getBirthdayViews(now time.Time, data []birthdayData) (today, upcoming, uncertain, err []birthdayData) {
	y, m, d := now.Date()
	nowDate := time.Date(y, m, d, 0, 0, 0, 0, time.Local)
	deadline := nowDate.AddDate(0, 0, 3)
	for _, v := range data {
		next := v.date.nextAfter(nowDate)
		switch {
			case v.error:
				err = append(err, v)
			case v.date.hasDay && v.date.hasMonth && deadline.After(next):
				if next.Day() == d && next.Month() == m {
					today = append(today, v)
				} else {
					upcoming = append(upcoming, v)
				}
			case !v.date.hasDay && v.date.hasMonth && next.Month() == m:
				uncertain = append(uncertain, v)
		}
	}
	return
}

// Construct messages from birthday data views for Show method
func getBirthdayMessages(today, upcoming, uncertain, err []birthdayData) (message, errMessage string) {
	var tempStrings, messageStrings []string
	
	if len(today) > 0 {
		tempStrings = []string{"Birthdays today:"}
		for _, v := range today {
			tempStrings = append(tempStrings, v.String())
		}
		messageStrings = append(messageStrings, strings.Join(tempStrings, "\n"))
	} else {
		messageStrings = append(messageStrings, "No birthdays today")
	}
	if len(upcoming) > 0 {
		tempStrings = []string{"Birthdays soon:"}
		for _, v := range upcoming {
			tempStrings = append(tempStrings, v.String())
		}
		messageStrings = append(messageStrings, strings.Join(tempStrings, "\n"))
	}
	if len(uncertain) > 0 {
		tempStrings = []string{"This month:"}
		for _, v := range uncertain {
			tempStrings = append(tempStrings, v.String())
		}
		messageStrings = append(messageStrings, strings.Join(tempStrings, "\n"))
	}
	
	if len(today) > 0 || len(upcoming) > 0 || len(uncertain) > 0 {
		message = strings.Join(messageStrings, "\n\n")
	}
	if len(err) > 0 {
		tempStrings = []string{"Unprocessed lines:"}
		for _, v := range err {
			tempStrings = append(tempStrings, v.raw)
		}
		errMessage = strings.Join(tempStrings, "\n")
	}
	return
}

// Show message box with actual/upcoming birthdays
func Show() {
	defer func(){
		if r := recover(); r != nil {
			ShowError(r)
		}
	}()
	data := load()
	today, upcoming, uncertain, err := getBirthdayViews(time.Now(), data)
	msg, errMsg := getBirthdayMessages(today, upcoming, uncertain, err)
	if len(msg) > 0 {
		messageBoxPlain("Birthdays reminder", msg)
	}
	if len(errMsg) > 0 {
		messageBoxPlain("ERRORS", errMsg)
	}
}

// Register OS-native task for automatic launch if needed
func RegisterHook() {
	defer func(){
		if r := recover(); r != nil {
			ShowError("Error on registration:", r)
		}
	}()
	if !isHookInstalled() {
		installHook()
	}
}
