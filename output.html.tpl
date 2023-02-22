<html>
  <title>ThiagosNews</title>
  <body>
    {{range .}}
    <a href="{{ .URL }}">{{ .Title }}</a><br />
    {{end}}
  </body>
</html>
