package rules

func GetDefaultRules() []Rule {
	return []Rule{
		NewDebugLogRule(),
		NewPlaintextPasswordRule(),
		NewBindAllRule(),
		NewTLSDisabledRule(),
		NewWeakAlgorithmRule(),
	}
}

func GetFileModeRules() []Rule {
	return append(GetDefaultRules(), NewFilePermissionRule())
}
