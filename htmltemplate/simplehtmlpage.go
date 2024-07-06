package htmltemplate

func GetSimpleHtmlMessagePage(simpleHtmlTitle string, simpleHtmlMessage string, err error) string {
	return `
			<!DOCTYPE html>
			<html>
			<head>
				<title>`+simpleHtmlTitle+`</title>
			</head>
			<body> <p> `+ simpleHtmlMessage +` <p> </body>
			</html>
			`
}