<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="utf-8">
    <meta http-equiv="X-UA-Compatible" content="IE=edge">
    <meta name="viewport" content="width=device-width, initial-scale=1">



    <title>IndieServ - {{.Name}}</title>
	<link rel="authorization_endpoint" href="{{AbsRoute "IndieAuthEndpoint"}}">
	<link rel="token_endpoint" href="{{AbsRoute "TokenEndpoint"}}">
	<link rel="micropub" href="{{AbsRoute "MicroPubEndpoint"}}">
	<link rel="webmention" href="{{AbsRoute "WebMentionEndpoint"}}">
	<link rel="hub" href="{{AbsRoute "PubHubEndpoint"}}">
	<link rel="self" href="{{AbsRoute "Post" "id" .Slug}}">
    <link href="/static/css/bootstrap.min.css" rel="stylesheet">
    <!--[if lt IE 9]>
      <script src="https://oss.maxcdn.com/html5shiv/3.7.2/html5shiv.min.js"></script>
      <script src="https://oss.maxcdn.com/respond/1.4.2/respond.min.js"></script>
    <![endif]-->
	<style>
		body {
			background-color: #F4F4F4;
		}
		.content {
			background-color: #fff;
			padding: 20px;
			border: 1px solid #ddd;
			margin-bottom: 10px;
		}
		.contentform {
			background-color: #eee;
			padding: 20px;
			border: 1px solid #ddd;
			margin-bottom: 10px;
		}
		.content > h1, .content > h2, .content > h3 {
			margin-top: 10px;
		}
		.empty-link {
			color: red;
		}
		.mention {
		    border-top: 1px solid #ddd;
		    margin-top: 5px;
		    padding-top: 5px;
		}
		.mention > .p-author,.mention > .p-summary {
		    margin-bottom: 0;
		}
	</style>
  </head>
  <body class="h-feed">
<nav class="navbar navbar-default">
  <div class="container-fluid">
    <div class="navbar-header">
      <button type="button" class="navbar-toggle collapsed" data-toggle="collapse" data-target="#bs-example-navbar-collapse-1">
        <span class="sr-only">Toggle navigation</span>
        <span class="icon-bar"></span>
        <span class="icon-bar"></span>
        <span class="icon-bar"></span>
      </button>
      <a class="navbar-brand" href="/">IndieServ</a>
    </div>

    <!-- Collect the nav links, forms, and other content for toggling -->
    <div class="collapse navbar-collapse" id="bs-example-navbar-collapse-1">
      <ul class="nav navbar-nav navbar-right">
		{{if .}}

		{{else}}

		{{end}}
		
      </ul>
    </div><!-- /.navbar-collapse -->
  </div><!-- /.container-fluid -->
</nav>
<div class="container">
<div class="row">
<div class="col-sm-3 h-card">
	<h3>Profile</h3>
	<a class="p-name u-url" href="{{.Profile.HomeURL}}">{{.Profile.Name}}</a>
	<a href="{{.Profile.Github}}" rel="me">Github</a>
</div>
<div class="col-sm-9">
<div class="content">
{{.Post}}
</div>
</div>

</div>
</div>
    <!-- jQuery (necessary for Bootstrap's JavaScript plugins) -->
    <script src="https://ajax.googleapis.com/ajax/libs/jquery/1.11.2/jquery.min.js"></script>
    <!-- Include all compiled plugins (below), or include individual files as needed -->
    <script src="/static/js/bootstrap.min.js"></script>
  </body>
</html>