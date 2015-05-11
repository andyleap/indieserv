<div class="form-group">
	<div class="col-sm-offset-2 col-sm-10">
    	<div class="checkbox">
			<label>
				<input type="checkbox" id="{{.Field.Var}}" name="{{.Field.Var}}"{{if .Value}} checked{{end}}>
				{{.Field.Name}}
			</label>
		</div>
	</div>
</div>