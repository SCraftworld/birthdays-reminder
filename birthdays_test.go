package birthdays

import (
	"fmt"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"testing"
	"time"
)

func TestNextAfter(t *testing.T) {
	d := time.Date(2019, time.Month(2), 2, 0, 0, 0, 0, time.Local)
	b := birthdayDate{"02.02.2019", time.Date(2019, time.Month(2), 2, 0, 0, 0, 0, time.Local), true, true, true}
	a := b.nextAfter(d)
	if !a.Equal(d) {
		t.Errorf("Bd today: expected %#v, actual %#v", d, a)
	}
	b = birthdayDate{"03.02.2019", time.Date(2019, time.Month(2), 3, 0, 0, 0, 0, time.Local), true, true, true}
	e := b.parsed
	a = b.nextAfter(d)
	if !a.Equal(e) {
		t.Errorf("Bd this year: expected %#v, actual %#v", e, a)
	}
	b = birthdayDate{"05.01.2019", time.Date(2019, time.Month(1), 5, 0, 0, 0, 0, time.Local), true, true, true}
	e = time.Date(2020, time.Month(1), 5, 0, 0, 0, 0, time.Local)
	a = b.nextAfter(d)
	if !a.Equal(e) {
		t.Errorf("Bd next year: expected %#v, actual %#v", e, a)
	}
	b = ParseBirthdayData("incorrectinput").date
	a = b.nextAfter(d)
	t.Log("Incorrect input single word: ok")
	
	b = ParseBirthdayData("incorrect input with spaces").date
	a = b.nextAfter(d)
	t.Log("Incorrect input multiple words: ok")
}

func assertBirthdayDateEqual(e, a birthdayDate, caseName string, t *testing.T) bool {
	if reflect.DeepEqual(e, a) { return true }
	t.Errorf("%s: expected '%#v', actual '%#v'", caseName, e, a)
	return false
}

func assertBirthdayDataEqual(e, a birthdayData, caseName string, t *testing.T) bool {
	if reflect.DeepEqual(e, a) { return true }
	t.Errorf("%s: expected '%#v', actual '%#v'", caseName, e, a)
	return false
}

func assertBirthdayDataStringEqual(e string, a birthdayData, caseName string, t *testing.T) bool {
	if e == a.String() { return true }
	t.Errorf("%s: expected '%s', actual '%s'", caseName, e, a.String())
	return false
}

func getAge(d birthdayDate) string {
	return strconv.Itoa(time.Now().Year() - d.parsed.Year())
}

func TestBirthdayDataParseAndString(t *testing.T) {
	s := "12.11.2019 test"
	d := birthdayDate{"12.11.2019", time.Date(2019, time.Month(11), 12, 0, 0, 0, 0, time.Local), true, true, true}
	e := birthdayData{s, "test", d, false}
	a := ParseBirthdayData(s)
	assertBirthdayDataEqual(e, a, "Simple parse", t)
	assertBirthdayDataStringEqual(fmt.Sprintf("%s - %s(%s)", e.name, e.date.raw, getAge(e.date)), a, "Simple string", t)
	s = "01.02.2019  test, тест, 02.03.2019"
	d = birthdayDate{"01.02.2019", time.Date(2019, time.Month(2), 1, 0, 0, 0, 0, time.Local), true, true, true}
	e = birthdayData{s, " test, тест, 02.03.2019", d, false}
	a = ParseBirthdayData(s)
	assertBirthdayDataEqual(e, a, "Complex name", t)
	assertBirthdayDataStringEqual(fmt.Sprintf("%s - %s(%s)", e.name, e.date.raw, getAge(e.date)), a, "Complex string", t)
	s = "??.02.???? test, тест, 02.03.2019"
	d = birthdayDate{"??.02.????", time.Date(1, time.Month(2), 1, 0, 0, 0, 0, time.Local), false, true, false}
	e = birthdayData{s, "test, тест, 02.03.2019", d, false}
	a = ParseBirthdayData(s)
	assertBirthdayDataEqual(e, a, "Missing date components", t)
	assertBirthdayDataStringEqual(fmt.Sprintf("%s - %s(?)", e.name, e.date.raw), a, "Missing age string", t)
	s = "incorrect"
	a = ParseBirthdayData(s)
	if !a.error || a.raw != s {
		t.Errorf("Incorrect input parse: actual '%#v'", a)
	}
	if a.error && a.raw != a.String() {
		t.Errorf("Incorrect input string: expected '%s', actual '%s'", a.raw, a.String())
	}
}

var sampleDataString string = `01.01.2012 te st 1
01.01.2012 test1-2
03.01.2011 test 3
//02.01.2012 comment

05.03.1990 test 4
01.12.1981 test 5
incorrect
??.01.???? mystery
1.5.3009 future
01.xy.1998 incorrect date
`

func sampleData() []birthdayData {
	return []birthdayData{ ParseBirthdayData("01.01.2012 te st 1"), // 0
						ParseBirthdayData("01.01.2012 test1-2"), // 1
						ParseBirthdayData("03.01.2011 test 3"), // 2
						ParseBirthdayData("05.03.1990 test 4"), // 3
						ParseBirthdayData("01.12.1981 test 5"), // 4
						ParseBirthdayData("incorrect"), // 5
						ParseBirthdayData("??.01.???? mystery"), // 6
						ParseBirthdayData("1.5.3009 future"), // 7
						ParseBirthdayData("01.xy.1998 incorrect date") } // 8
}

func TestLoad(t *testing.T) {
	filename := "BD.txt"
	if err := os.WriteFile(filename, []byte(sampleDataString), 0); err != nil {
		t.Fatal("Can't prepare file", err)
	}
	defer func() {
		if err := os.Remove(filename); err != nil {
			t.Error("Can't remove file", err)
		}
	}()
	a := load()
	e := sampleData()
	if !reflect.DeepEqual(a, e) {
		t.Errorf("Incorrect data loaded, expected '%#v', actual '%#v'", e, a)
	}
}

func assertViewsEqual(date time.Time, data, todayE, upcomingE, uncertainE, errE []birthdayData, caseName string, t *testing.T) {
	todayA, upcomingA, uncertainA, errA := getBirthdayViews(date, data)
	if !reflect.DeepEqual(todayA, todayE) {
		t.Errorf("%s 'today' mismatch: expected '%#v', actual '%#v'", caseName, todayE, todayA)
	}
	if !reflect.DeepEqual(upcomingA, upcomingE) {
		t.Errorf("%s 'upcoming' mismatch: expected '%#v', actual '%#v'", caseName, upcomingE, upcomingA)
	}
	if !reflect.DeepEqual(uncertainA, uncertainE) {
		t.Errorf("%s 'uncertain' mismatch: expected '%#v', actual '%#v'", caseName, uncertainE, uncertainA)
	}
	if !reflect.DeepEqual(errA, errE) {
		t.Errorf("%s 'err' mismatch: expected '%#v', actual '%#v'", caseName, errE, errA)
	}
}

func TestGetBirthdayViews(t *testing.T) {
	data := sampleData()
	todayE := []birthdayData{data[0], data[1]}
	upcomingE := []birthdayData{data[2]}
	uncertainE := []birthdayData{data[6]}
	errE := []birthdayData{data[5], data[8]} //unchanged through all cases
	assertViewsEqual(time.Date(2020, time.Month(1), 1, 0, 0, 0, 0, time.Local), data, todayE, upcomingE, uncertainE, errE, "Single day multiple events", t)
	
	todayE = nil
	upcomingE = nil
	uncertainE = nil
	assertViewsEqual(time.Date(2020, time.Month(2), 4, 0, 0, 0, 0, time.Local), data, todayE, upcomingE, uncertainE, errE, "Single day no events", t)
	
	todayE = nil
	upcomingE = []birthdayData{data[0], data[1]}
	uncertainE = nil
	assertViewsEqual(time.Date(2020, time.Month(12), 31, 0, 0, 0, 0, time.Local), data, todayE, upcomingE, uncertainE, errE, "Year border", t)
	
	todayE = []birthdayData{data[7]}
	upcomingE = nil
	uncertainE = nil
	assertViewsEqual(time.Date(2020, time.Month(5), 1, 0, 0, 0, 0, time.Local), data, todayE, upcomingE, uncertainE, errE, "Future", t)
}

func TestGetBirthdayMessages(t *testing.T) {
	var messageE, errMessageE string
	var today, upcoming, uncertain, err []birthdayData
	
	assertMessagesEqual := func(caseName string) {
		messageA, errMessageA := getBirthdayMessages(today, upcoming, uncertain, err)
		re := regexp.MustCompile(`\(\d+\)`)
		messageA = re.ReplaceAllString(messageA, `(xx)`)
		if messageA != messageE {
			t.Errorf("%s message mismatch: expected '%s', actual '%s'", caseName, messageE, messageA)
		}
		if errMessageA != errMessageE {
			t.Errorf("%s error message mismatch: expected '%s', actual '%s'", caseName, errMessageE, errMessageA)
		}
	}
	
	data := sampleData()
	today = []birthdayData{data[0], data[2]}
	upcoming = []birthdayData{data[3]}
	uncertain = []birthdayData{data[4]}
	err = []birthdayData{data[5], data[8]}
	messageE = `Сегодня празднуют:
te st 1 - 01.01.2012(xx)
test 3 - 03.01.2011(xx)

Скоро празднуют:
test 4 - 05.03.1990(xx)

В этом месяце:
test 5 - 01.12.1981(xx)`
	errMessageE = `Не обработаны строки:
incorrect
01.xy.1998 incorrect date`
	assertMessagesEqual("All fields")
	
	today = nil
	upcoming = nil
	uncertain = nil
	err = nil
	messageE = ``
	errMessageE = ``
	assertMessagesEqual("No fields")
	
	today = nil
	upcoming = []birthdayData{data[3]}
	err = nil
	messageE = `Сегодня нет именинников

Скоро празднуют:
test 4 - 05.03.1990(xx)`
	errMessageE = ``
	assertMessagesEqual("No events today but upcoming present")
}

func TestRegisterHook(t *testing.T) {
	if !isAdmin() {
		t.Skip("Not admin")
		return
	}
	if isHookInstalled() {
		uninstallHook()
		if isHookInstalled() {
			t.Fatal("Can't remove existing hook.")
		}
	}
	
	installHook()
	if !isHookInstalled() {
		t.Fatal("Can't install hook")
	}
}
