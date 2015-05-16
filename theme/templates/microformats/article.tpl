<div class="h-entry">
	<h1 class="p-name"><a href="{{AbsRoute "Post" "id" .Slug}}" class="u-url">{{.Title}}</a></h1>
	<span class="e-content p-name">
		{{.Content | SafeHTML}}
	</span>
	<div class="publish-info">
		<time class="dt-published" datetime="{{.Published.Format "2006-01-02T15:04:05Z07:00"}}">{{.Published.Format "Jan 2, 2006 at 3:04pm"}}</time>
		<a class="u-url" href="{{AbsRoute "Post" "id" .Slug}}">{{AbsRoute "Post" "id" .Slug}}</a>
	</div>
	{{range .MentionItems}}
	{{template "mention.tpl" .}}
	{{end}}
</div>