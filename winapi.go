// +build windows
package birthdays

import (
	"bytes"
	"golang.org/x/sys/windows"
	"golang.org/x/text/encoding/charmap"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"text/template"
	"unicode/utf16"
	"unicode/utf8"
)

import _ "embed"

// Convert utf8 string to utf16 pointer for Win32 API consumption
func strToUTF16PtrWithPanic(s string) *uint16 {
	res, err := syscall.UTF16PtrFromString(s)
	if err != nil {
		panic(err)
	}
	return res
}

// MessageBox of Win32 API
func messageBoxPlain(caption, text string) {
	_, err := windows.MessageBox(
		0,
		strToUTF16PtrWithPanic(text),
		strToUTF16PtrWithPanic(caption),
		0)
	if err != nil {
		panic(err)
	}
}

// Check if current process running with admin privileges
func isAdmin() bool {
	var sid *windows.SID
	err := windows.AllocateAndInitializeSid(
			&windows.SECURITY_NT_AUTHORITY,
			2,
			windows.SECURITY_BUILTIN_DOMAIN_RID,
			windows.DOMAIN_ALIAS_RID_ADMINS,
			0, 0, 0, 0, 0, 0,
			&sid)
	if err != nil {
			panic(err)
	}
	defer windows.FreeSid(sid)
	token := windows.Token(0)
	member, err := token.IsMember(sid)
	if err != nil {
			panic(err)
	}

	return member
}

// Start current executable with same arguments with admin privileges and finish non-privileged process
func restartAsAdmin() {
	verb := "runas"
    exe, _ := os.Executable()
    cwd, _ := os.Getwd()
    args := strings.Join(os.Args[1:], " ")
	
    verbPtr := strToUTF16PtrWithPanic(verb)
    exePtr := strToUTF16PtrWithPanic(exe)
    cwdPtr := strToUTF16PtrWithPanic(cwd)
    argPtr := strToUTF16PtrWithPanic(args)

    var showCmd int32 = 1 //SW_NORMAL

    if err := windows.ShellExecute(0, verbPtr, exePtr, argPtr, cwdPtr, showCmd); err != nil {
        panic(err)
    } else {
		os.Exit(1)
	}
}

// Check if OS-native task for automatic launch already present
func isHookInstalled() bool {
	return (exec.Command(`schtasks`, `/query`, `/tn`, `scraftworld\birthdays`).Run() == nil)
}

//go:embed resources/birthdays.xml
var taskBytes []byte

type TaskInfo struct {
	ExePath string
	WorkingDirectory string
}

// Convert utf16 byte slice to utf8 string
func decodeUTF16(b []byte) string {
	u16s := make([]uint16, 1)
	ret := &bytes.Buffer{}
	b8buf := make([]byte, 4)
	lb := 2*(len(b)/2)
	for i := 0; i < lb; i += 2 {
		u16s[0] = uint16(b[i]) + (uint16(b[i+1]) << 8)
		r := utf16.Decode(u16s)
		n := utf8.EncodeRune(b8buf, r[0])
		ret.Write(b8buf[:n])
	}
	return ret.String()
}

// Convert utf8 string to utf16 byte slice
func encodeUTF16(s string) []byte {
	cps := utf16.Encode([]rune(s))
	res := make([]byte, len(cps) * 2)
	for i, cp := range cps {
		res[i*2] = byte(cp)
		res[i*2+1] = byte(cp>>8)
	}
	return res
}

// Convert utf16 xml task config to utf8 for templating and back
func templateTaskXml() (res []byte, err error) {
	exePath, err := os.Executable()
	if err != nil { return }
	taskInfo := TaskInfo{exePath, filepath.Dir(exePath)}
	taskXml := decodeUTF16(taskBytes)
	tmpl, err := template.New("taskTemplate").Parse(taskXml)
	if err != nil { return }
	var resBuilder strings.Builder
	err = tmpl.Execute(&resBuilder, taskInfo)
	if err != nil { return }
	res = encodeUTF16(resBuilder.String())
	return
}

var taskName = `scraftworld\birthdays`

// Register OS-native task for automatic launch
func installHook() {
	if !isAdmin() {
		restartAsAdmin()
		return
	}
	taskInfo, err := templateTaskXml()
	if err != nil { panic(err) }
	
	taskFilename := "birthdays.xml"
	taskFile, err := os.Create(taskFilename)
	if err != nil { panic(err) }
	
	_, err = taskFile.Write(taskInfo)
	taskFile.Close()
	if err != nil { panic(err) }
	defer func(){
		os.Remove(taskFilename)
	}()
	cmd := exec.Command(`schtasks`, `/create`, `/tn`, taskName, `/xml`, taskFilename)
	if output, err := cmd.CombinedOutput(); err != nil {
		output, err := charmap.CodePage866.NewDecoder().Bytes(output)
		if err != nil {
			output = []byte("<can't decode schtasks output>")
		}
		panic(string(output))
	}
}

func uninstallHook() {
	if !isAdmin() {
		restartAsAdmin()
		return
	}
	cmd := exec.Command(`schtasks`, `/delete`, `/tn`, taskName)
	if output, err := cmd.CombinedOutput(); err != nil {
		output, err := charmap.CodePage866.NewDecoder().Bytes(output)
		if err != nil {
			output = []byte("<can't decode schtasks output>")
		}
		panic(string(output))
	}
}
