package utils

import (
	"golang.org/x/sys/windows"
)

// IsAdministrator 判断是否在管理员模式下运行
func IsAdministrator() bool {
	var sid *windows.SID
	err := windows.AllocateAndInitializeSid(&windows.SECURITY_NT_AUTHORITY,
		2, windows.SECURITY_BUILTIN_DOMAIN_RID, windows.DOMAIN_ALIAS_RID_ADMINS,
		0, 0, 0, 0, 0, 0, &sid)
	if err != nil {
		return false
	}

	defer windows.FreeSid(sid)

	var token windows.Token

	//defer token.Close()
	isMember, err := token.IsMember(sid)
	if err != nil {
		return false
	}

	return isMember
}
