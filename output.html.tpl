<html lang="en">
  <head>
    <title>ThiagosNews</title>
    <link href="https://maxcdn.bootstrapcdn.com/bootstrap/latest/css/bootstrap.min.css" rel="stylesheet">
  </head>
  <body>
    <div class="list-group">
      {{- range . }}
      <li class="list-group-item"><a href="{{ .URL }}">{{ .Title }}</a></li>
      {{- end }}
    </div>
  </body>
</html>
