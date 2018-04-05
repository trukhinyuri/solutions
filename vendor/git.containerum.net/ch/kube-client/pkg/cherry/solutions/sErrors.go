package sErrors

import (
	bytes "bytes"
	cherry "git.containerum.net/ch/kube-client/pkg/cherry"
	template "text/template"
)

const ()

func ErrAdminRequired(params ...func(*cherry.Err)) *cherry.Err {
	err := &cherry.Err{Message: "Admin access required", StatusHTTP: 403, ID: cherry.ErrID{SID: 0xb, Kind: 0x1}, Details: []string(nil)}
	for _, param := range params {
		param(err)
	}
	for i, detail := range err.Details {
		det := renderTemplate(detail)
		err.Details[i] = det
	}
	return err
}

func ErrRequiredHeadersNotProvided(params ...func(*cherry.Err)) *cherry.Err {
	err := &cherry.Err{Message: "Required headers not provided", StatusHTTP: 400, ID: cherry.ErrID{SID: 0xb, Kind: 0x2}, Details: []string(nil)}
	for _, param := range params {
		param(err)
	}
	for i, detail := range err.Details {
		det := renderTemplate(detail)
		err.Details[i] = det
	}
	return err
}

func ErrRequestValidationFailed(params ...func(*cherry.Err)) *cherry.Err {
	err := &cherry.Err{Message: "Request validation failed", StatusHTTP: 400, ID: cherry.ErrID{SID: 0xb, Kind: 0x3}, Details: []string(nil)}
	for _, param := range params {
		param(err)
	}
	for i, detail := range err.Details {
		det := renderTemplate(detail)
		err.Details[i] = det
	}
	return err
}

func ErrUnableUpdateSolutionsList(params ...func(*cherry.Err)) *cherry.Err {
	err := &cherry.Err{Message: "Unable to update solutions template list", StatusHTTP: 500, ID: cherry.ErrID{SID: 0xb, Kind: 0x4}, Details: []string(nil)}
	for _, param := range params {
		param(err)
	}
	for i, detail := range err.Details {
		det := renderTemplate(detail)
		err.Details[i] = det
	}
	return err
}

func ErrUnableGetSolutionsTemplatesList(params ...func(*cherry.Err)) *cherry.Err {
	err := &cherry.Err{Message: "Unable to get solutions templates list", StatusHTTP: 500, ID: cherry.ErrID{SID: 0xb, Kind: 0x5}, Details: []string(nil)}
	for _, param := range params {
		param(err)
	}
	for i, detail := range err.Details {
		det := renderTemplate(detail)
		err.Details[i] = det
	}
	return err
}

func ErrUnableGetSolutionTemplate(params ...func(*cherry.Err)) *cherry.Err {
	err := &cherry.Err{Message: "Unable to get solutions template", StatusHTTP: 500, ID: cherry.ErrID{SID: 0xb, Kind: 0x6}, Details: []string(nil)}
	for _, param := range params {
		param(err)
	}
	for i, detail := range err.Details {
		det := renderTemplate(detail)
		err.Details[i] = det
	}
	return err
}

func ErrUnableGetSolutionsList(params ...func(*cherry.Err)) *cherry.Err {
	err := &cherry.Err{Message: "Unable to get user solutions list", StatusHTTP: 500, ID: cherry.ErrID{SID: 0xb, Kind: 0x7}, Details: []string(nil)}
	for _, param := range params {
		param(err)
	}
	for i, detail := range err.Details {
		det := renderTemplate(detail)
		err.Details[i] = det
	}
	return err
}

func ErrUnableGetSolution(params ...func(*cherry.Err)) *cherry.Err {
	err := &cherry.Err{Message: "Unable to get user solution", StatusHTTP: 500, ID: cherry.ErrID{SID: 0xb, Kind: 0x8}, Details: []string(nil)}
	for _, param := range params {
		param(err)
	}
	for i, detail := range err.Details {
		det := renderTemplate(detail)
		err.Details[i] = det
	}
	return err
}

func ErrUnableCreateSolution(params ...func(*cherry.Err)) *cherry.Err {
	err := &cherry.Err{Message: "Unable to create solution", StatusHTTP: 500, ID: cherry.ErrID{SID: 0xb, Kind: 0x9}, Details: []string(nil)}
	for _, param := range params {
		param(err)
	}
	for i, detail := range err.Details {
		det := renderTemplate(detail)
		err.Details[i] = det
	}
	return err
}

func ErrUnableDeleteSolution(params ...func(*cherry.Err)) *cherry.Err {
	err := &cherry.Err{Message: "Unable to delete solution", StatusHTTP: 500, ID: cherry.ErrID{SID: 0xb, Kind: 0xa}, Details: []string(nil)}
	for _, param := range params {
		param(err)
	}
	for i, detail := range err.Details {
		det := renderTemplate(detail)
		err.Details[i] = det
	}
	return err
}

func ErrSolutionAlreadyExists(params ...func(*cherry.Err)) *cherry.Err {
	err := &cherry.Err{Message: "Solution with this name already exists", StatusHTTP: 409, ID: cherry.ErrID{SID: 0xb, Kind: 0xb}, Details: []string(nil)}
	for _, param := range params {
		param(err)
	}
	for i, detail := range err.Details {
		det := renderTemplate(detail)
		err.Details[i] = det
	}
	return err
}

func ErrSolutionNotExist(params ...func(*cherry.Err)) *cherry.Err {
	err := &cherry.Err{Message: "Solution with this name doesn't exist", StatusHTTP: 404, ID: cherry.ErrID{SID: 0xb, Kind: 0xc}, Details: []string(nil)}
	for _, param := range params {
		param(err)
	}
	for i, detail := range err.Details {
		det := renderTemplate(detail)
		err.Details[i] = det
	}
	return err
}

func ErrInternalError(params ...func(*cherry.Err)) *cherry.Err {
	err := &cherry.Err{Message: "Internal error", StatusHTTP: 500, ID: cherry.ErrID{SID: 0xb, Kind: 0xd}, Details: []string(nil)}
	for _, param := range params {
		param(err)
	}
	for i, detail := range err.Details {
		det := renderTemplate(detail)
		err.Details[i] = det
	}
	return err
}
func renderTemplate(templText string) string {
	buf := &bytes.Buffer{}
	templ, err := template.New("").Parse(templText)
	if err != nil {
		return err.Error()
	}
	err = templ.Execute(buf, map[string]string{})
	if err != nil {
		return err.Error()
	}
	return buf.String()
}
