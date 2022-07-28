package app

import "dsh/px/db"

// NewContext creates and returns [Context] for appID
func (self *Global) NewContext(appID string) (*Context, error) {
	db, err := self.db.DB(appID)
	if err != nil {
		return nil, err
	}

	return &Context{
		appID: appID,
		db:    db,
	}, nil
}

// Context defines local app context for current request. We are creating it in
// the beginning of the request and working with it during processing of the
// request.
type Context struct {
	// ID of application this context for. Also we use it as DB name for this app.
	appID string

	// DB pools, initialized to connect to DB with name appID.
	db *db.DB
}

// AppID returns ID of app this [Context] was created for
func (self *Context) AppID() string {
	return self.appID
}
