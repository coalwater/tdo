package cmd

import "strings"

// extractFlags pulls --id, --json, and --help from raw args when
// DisableFlagParsing is true. Returns extracted values + remaining args.
// Stops extracting after "--" (passes it and everything after through to remaining).
func extractFlags(args []string) (id string, json bool, help bool, remaining []string) {
	i := 0
	for i < len(args) {
		arg := args[i]
		if arg == "--" {
			remaining = append(remaining, args[i:]...)
			break
		}
		switch {
		case arg == "--id" && i+1 < len(args):
			id = args[i+1]
			i += 2
		case strings.HasPrefix(arg, "--id="):
			id = strings.TrimPrefix(arg, "--id=")
			i++
		case arg == "--json":
			json = true
			i++
		case arg == "--help" || arg == "-h":
			help = true
			i++
		default:
			remaining = append(remaining, arg)
			i++
		}
	}
	return
}
