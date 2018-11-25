package auth0

import (
	"encoding/json"
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	auth0 "github.com/yieldr/go-auth0"
	"github.com/yieldr/go-auth0/management"
)

func newUser() *schema.Resource {
	return &schema.Resource{
		Create: createUser,
		Read:   readUser,
		Update: updateUser,
		Delete: deleteUser,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"user_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return old == "auth0|"+new
				},
			},
			"connection_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"email": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"username": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"password": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"phone_number": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"user_metadata": {
				Type:     schema.TypeMap,
				Optional: true,
			},
			"email_verified": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"verify_email": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"phone_verified": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"app_metadata": {
				Type:     schema.TypeMap,
				Optional: true,
				ConflictsWith: []string{"app_metadata_raw"},
			},
			"app_metadata_raw": {
				Type: schema.TypeString,
				Optional: true,
				ConflictsWith: []string{"app_metadata"},
				ValidateFunc: func(i interface{}, s string) ([]string,[]error) {
					var errs []error
					iStr, ok := i.(string)
					if !ok {
						errs = append(errs, fmt.Errorf("%q must be a string", s))
					}

					var decoded map[string]interface{}
					err := json.Unmarshal([]byte(iStr), &decoded)

					if err != nil {
						errs = append(errs, fmt.Errorf("invalid json in %q. %v", s, err.Error()))
					}

					return nil, errs
				},
			},
		},
	}
}

func readUser(d *schema.ResourceData, m interface{}) error {
	api := m.(*management.Management)
	u, err := api.User.Read(d.Id())
	if err != nil {
		return err
	}
	d.Set("user_id", u.ID)
	d.Set("username", u.Username)
	d.Set("phone_number", u.PhoneNumber)
	d.Set("user_metadata", u.UserMetadata)
	d.Set("email_verified", u.EmailVerified)
	d.Set("phone_verified", u.PhoneVerified)
	d.Set("verify_email", u.VerifyEmail)
	d.Set("app_metadata", u.AppMetadata)
	d.Set("email", u.Email)
	return nil
}

func createUser(d *schema.ResourceData, m interface{}) error {
	u := buildUser(d)
	api := m.(*management.Management)
	if err := api.User.Create(u); err != nil {
		return err
	}
	d.SetId(*u.ID)
	return nil
}

func updateUser(d *schema.ResourceData, m interface{}) error {
	u := buildUser(d)
	api := m.(*management.Management)
	if err := api.User.Update(d.Id(), u); err != nil {
		return err
	}
	return readUser(d, m)
}

func deleteUser(d *schema.ResourceData, m interface{}) error {
	api := m.(*management.Management)
	return api.User.Delete(d.Id())
}

func buildUser(d *schema.ResourceData) *management.User {
	var appMetadata map[string]interface{}
	
	if val, ok := d.GetOk("app_metadata_raw"); ok {
		json.Unmarshal([]byte(val.(string)), &appMetadata)
	} else {
		appMetadata = Map(d, "app_metadata")
	}
	 
	u := &management.User{
		ID:            String(d, "user_id"),
		Connection:    String(d, "connection_name"),
		Username:      String(d, "username"),
		PhoneNumber:   String(d, "phone_number"),
		UserMetadata:  Map(d, "user_metadata"),
		EmailVerified: Bool(d, "email_verified"),
		VerifyEmail:   Bool(d, "verify_email"),
		PhoneVerified: Bool(d, "phone_verified"),
		AppMetadata:   appMetadata,
		Email:         String(d, "email"),
		Password:      String(d, "password"),
	}

	if u.Username != nil || u.Password != nil || u.EmailVerified != nil || u.PhoneVerified != nil {
		// When updating email_verified, phone_verified, username or password
		// we need to specify the connection property too.
		//
		// https://auth0.com/docs/api/management/v2#!/Users/patch_users_by_id
		//
		// As the builtin String function internally checks if the key has been
		// changed, we retrieve the value of "connection_name" regardless of
		// change.
		u.Connection = auth0.String(d.Get("connection_name").(string))
	}

	return u
}
