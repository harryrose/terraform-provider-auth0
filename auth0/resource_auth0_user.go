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
				Type:          schema.TypeMap,
				Optional:      true,
				ConflictsWith: []string{"user_metadata_json"},
				ValidateFunc:  validateJSONField,
			},
			"user_metadata_json": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"user_metadata"},
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
				Type:          schema.TypeMap,
				Optional:      true,
				ConflictsWith: []string{"app_metadata_json"},
				ValidateFunc:  validateJSONField,
			},
			"app_metadata_json": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"app_metadata"},
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

	if js, err := json.Marshal(u.UserMetadata); err != nil {
		d.Set("user_metadata_json", string(js))
	}

	d.Set("email_verified", u.EmailVerified)
	d.Set("phone_verified", u.PhoneVerified)
	d.Set("verify_email", u.VerifyEmail)
	d.Set("app_metadata", u.AppMetadata)

	if js, err := json.Marshal(u.AppMetadata); err != nil {
		d.Set("app_metadata_json", string(js))
	}

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

	appMetadata := handleJSONFields(d, "app_metadata_json", "app_metadata")
	userMetadata := handleJSONFields(d, "user_metadata_json", "user_metadata")

	u := &management.User{
		ID:            String(d, "user_id"),
		Connection:    String(d, "connection_name"),
		Username:      String(d, "username"),
		PhoneNumber:   String(d, "phone_number"),
		UserMetadata:  userMetadata,
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

func handleJSONFields(d *schema.ResourceData, jsonFieldName, mapFieldName string) map[string]interface{} {
	m := Map(d, mapFieldName)
	if m != nil {
		return m
	}

	js := String(d, jsonFieldName)
	if js == nil {
		return nil
	}

	_ = json.Unmarshal([]byte(*js), &m)

	return m
}

func validateJSONField(value interface{}, key string) (strings []string, errors []error) {
	str, ok := value.(string)

	if !ok {
		errors = append(errors, fmt.Errorf("field '%s' must be a string", key))
		return
	}

	var target map[string]interface{}

	err := json.Unmarshal([]byte(str), &target)

	if err != nil {
		errors = append(errors, fmt.Errorf("field '%s' is not a valid JSON object. %v", key, err))
	}
	return
}
