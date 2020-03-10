package main

import (
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/jinzhu/now"
	"github.com/qor/auth/auth_identity"
	"github.com/qor/auth/providers/password"
	"github.com/qor/banner_editor"
	"github.com/qor/help"
	i18n_database "github.com/qor/i18n/backends/database"
	"github.com/qor/media/asset_manager"
	"github.com/qor/notification"
	"github.com/qor/notification/channels/database"
	"github.com/qor/qor"
	"github.com/qor/seo"
	qortenant_database "github.com/snowlyg/qortenant/backend/database"
	"go-tenancy/config/auth"
	"go-tenancy/config/db"
	_ "go-tenancy/config/db/migrations"
	adminseo "go-tenancy/models/seo"
	"go-tenancy/models/settings"
	"go-tenancy/models/stores"
	"go-tenancy/models/users"
	"go-tenancy/utils"
)

/* How to run this script
   $ go run db/seeds/main.go db/seeds/seeds.go
*/

/* How to upload file
 * $ brew install s3cmd
 * $ s3cmd --configure (Refer https://github.com/theplant/qor-example)
 * $ s3cmd put local_file_path s3://qor3/
 */

var (
	AdminUser    *users.User
	Notification = notification.New(&notification.Config{})
	Tables       = []interface{}{
		&auth_identity.AuthIdentity{},
		&users.User{}, &users.Address{},
		&stores.Store{},
		&settings.Setting{},
		&adminseo.MySEOSetting{},
		&settings.MediaLibrary{},
		&banner_editor.QorBannerEditorSetting{},

		&asset_manager.AssetManager{},
		&i18n_database.Translation{},
		&notification.QorNotification{},

		&help.QorHelpEntry{},

		&qortenant_database.Tenant{},
		&qortenant_database.TUser{},
		&qortenant_database.TOauthToken{},
		&qortenant_database.TRole{},
		&qortenant_database.TPermission{},
	}
)

func main() {
	Notification.RegisterChannel(database.New(&database.Config{}))
	TruncateTables(Tables...)
	createRecords()
}

func createRecords() {
	fmt.Println("开始填充数据...")

	createSetting()
	fmt.Println("--> 填充 setting.")

	createSeo()
	fmt.Println("--> 填充 seo.")

	createAdminUsers()
	fmt.Println("--> 填充 admin users.")

	createUsers()
	fmt.Println("--> 填充 users.")

	createAddresses()
	fmt.Println("--> 填充 addresses.")

	createTPermissions()
	fmt.Println("--> 填充 RabcPermission.")

	createTRoles()
	fmt.Println("--> 填充 RabcRole.")

	createTUsers()
	fmt.Println("--> 填充 RabcUser.")

	createTenants()
	fmt.Println("--> 填充 tenants.")

	createMediaLibraries()
	fmt.Println("--> 填充 medialibraries.")

	createStores()
	fmt.Println("--> 填充 stores.")

	createHelps()
	fmt.Println("--> 填充 helps.")

	fmt.Println("--> 完成!")
}

func createSetting() {
	setting := settings.Setting{}

	setting.ShippingFee = Seeds.Setting.ShippingFee
	setting.GiftWrappingFee = Seeds.Setting.GiftWrappingFee
	setting.CODFee = Seeds.Setting.CODFee
	setting.TaxRate = Seeds.Setting.TaxRate
	setting.Address = Seeds.Setting.Address
	setting.Region = Seeds.Setting.Region
	setting.City = Seeds.Setting.City
	setting.Country = Seeds.Setting.Country
	setting.Zip = Seeds.Setting.Zip
	setting.Latitude = Seeds.Setting.Latitude
	setting.Longitude = Seeds.Setting.Longitude

	if err := DraftDB.Create(&setting).Error; err != nil {
		fmt.Println(fmt.Sprintf("create setting (%v) failure, got err %v", setting, err))
	}
}

func createSeo() {
	globalSeoSetting := adminseo.MySEOSetting{}
	globalSetting := make(map[string]string)
	globalSetting["SiteName"] = "Qor Demo"
	globalSeoSetting.Setting = seo.Setting{GlobalSetting: globalSetting}
	globalSeoSetting.Name = "QorSeoGlobalSettings"
	globalSeoSetting.LanguageCode = "en-US"
	globalSeoSetting.QorSEOSetting.SetIsGlobalSEO(true)

	if err := db.DB.Create(&globalSeoSetting).Error; err != nil {
		fmt.Println(fmt.Sprintf("create seo (%v) failure, got err %v", globalSeoSetting, err))
	}

	defaultSeo := adminseo.MySEOSetting{}
	defaultSeo.Setting = seo.Setting{Title: "{{SiteName}}", Description: "{{SiteName}} - Default Description", Keywords: "{{SiteName}} - Default Keywords", Type: "Default Page"}
	defaultSeo.Name = "Default Page"
	defaultSeo.LanguageCode = "en-US"
	if err := db.DB.Create(&defaultSeo).Error; err != nil {
		fmt.Println(fmt.Sprintf("create seo (%v) failure, got err %v", defaultSeo, err))
	}
}

func createAdminUsers() {
	AdminUser = &users.User{}
	AdminUser.Email = "dev@getqor.com"
	AdminUser.Confirmed = true
	AdminUser.Name = "SuperAdmin"
	AdminUser.Role = "SuperAdmin"
	if avatar, err := os.Open("config/db/seeds/data/avatars/2.jpg"); err != nil {
		panic(fmt.Sprintf("file doesn't exist %v\n", err))
	} else {
		_ = AdminUser.Avatar.Scan(avatar)
	}

	DraftDB.Create(AdminUser)

	provider := auth.Auth.GetProvider("password").(*password.Provider)
	hashedPassword, err := provider.Encryptor.Digest("testing")
	if err != nil {
		log.Fatal(fmt.Sprintf(" provider.Encryptor.Digest error %v", err))
	}

	now := time.Now()

	authIdentity := &auth_identity.AuthIdentity{}
	authIdentity.Provider = "password"
	authIdentity.UID = AdminUser.Email
	authIdentity.EncryptedPassword = hashedPassword
	authIdentity.UserID = fmt.Sprint(AdminUser.ID)
	authIdentity.ConfirmedAt = &now

	DraftDB.Create(authIdentity)

	// Send welcome notification
	_ = Notification.Send(&notification.Message{
		From:        AdminUser,
		To:          AdminUser,
		Title:       "Welcome To QOR Admin",
		Body:        "Welcome To QOR Admin",
		MessageType: "info",
	}, &qor.Context{DB: DraftDB})
}

func createUsers() {
	emailRegexp := regexp.MustCompile(".*(@.*)")
	totalCount := 600
	for i := 0; i < totalCount; i++ {
		user := users.User{}
		user.Name = Fake.Name()
		user.Email = emailRegexp.ReplaceAllString(Fake.Email(), strings.Replace(strings.ToLower(user.Name), " ", "_", -1)+"@example.com")
		user.Gender = []string{"Female", "Male"}[i%2]
		if err := DraftDB.Create(&user).Error; err != nil {
			log.Fatal(fmt.Sprintf("create user (%v) failure, got err %v", user, err))
		}

		day := -14 + i/45
		user.CreatedAt = now.EndOfDay().Add(time.Duration(day*rand.Intn(24)) * time.Hour)
		if user.CreatedAt.After(time.Now()) {
			user.CreatedAt = time.Now()
		}
		if err := DraftDB.Save(&user).Error; err != nil {
			log.Fatal(fmt.Sprintf("Save user (%v) failure, got err %v", user, err))
		}

		provider := auth.Auth.GetProvider("password").(*password.Provider)
		hashedPassword, err := provider.Encryptor.Digest("testing")
		if err != nil {
			log.Fatal(fmt.Sprintf(" provider.Encryptor.Digest error %v", err))
		}
		authIdentity := &auth_identity.AuthIdentity{}
		authIdentity.Provider = "password"
		authIdentity.UID = user.Email
		authIdentity.EncryptedPassword = hashedPassword
		authIdentity.UserID = fmt.Sprint(user.ID)
		authIdentity.ConfirmedAt = &user.CreatedAt

		DraftDB.Create(authIdentity)
	}
}

func createAddresses() {
	var Users []users.User
	if err := DraftDB.Find(&Users).Error; err != nil {
		log.Fatal(fmt.Sprintf("query users (%v) failure, got err %v", Users, err))
	}

	for _, user := range Users {
		address := users.Address{}
		address.UserID = user.ID
		address.ContactName = user.Name
		address.Phone = Fake.PhoneNumber()
		address.City = Fake.City()
		address.Address1 = Fake.StreetAddress()
		address.Address2 = Fake.SecondaryAddress()
		if err := DraftDB.Create(&address).Error; err != nil {
			log.Fatal(fmt.Sprintf("create address (%v) failure, got err %v", address, err))
		}
	}
}

func createTPermissions() {
	for _, u := range Seeds.TPermissions {
		tpermission := qortenant_database.TPermission{}
		tpermission.Name = u.Name
		tpermission.DisplayName = u.DisplayName
		tpermission.Description = u.Description
		tpermission.Act = u.Act
		if err := tpermission.CreateTPermission(); err != nil {
			log.Fatal(fmt.Sprintf("create TPermission (%v) failure, got err %v", tpermission, err))
		}
	}
}

func createTRoles() {

	var rabcPermIds []uint
	rabcPerms := qortenant_database.GetAllTPermissions("", "", 0, 0)
	for _, rabcPerm := range rabcPerms {
		rabcPermIds = append(rabcPermIds, rabcPerm.ID)
	}

	for _, u := range Seeds.TRoles {
		trole := qortenant_database.TRole{}
		trole.Name = u.Name
		trole.DisplayName = u.DisplayName
		trole.Description = u.Description
		if err := trole.CreateTRole(rabcPermIds); err != nil {
			log.Fatal(fmt.Sprintf("create TRole (%v) failure, got err %v", trole, err))
		}
	}
}

func createTUsers() {
	var rabcRoleIds []uint
	rabcRoles := qortenant_database.GetAllTRoles("", "", 0, 0)
	for _, rabcRole := range rabcRoles {
		rabcRoleIds = append(rabcRoleIds, rabcRole.ID)
	}

	for _, u := range Seeds.TUsers {
		tuser := qortenant_database.TUser{}
		tuser.Name = u.Name
		tuser.Username = u.Username
		tuser.Password = utils.HashPassword("password")
		if err := tuser.CreateTUser(rabcRoleIds); err != nil {
			log.Fatal(fmt.Sprintf("create TUser (%v) failure, got err %v", tuser, err))
		}
	}
}

func createTenants() {
	for _, t := range Seeds.Tenants {
		tt := qortenant_database.Tenant{}
		tt.Name = t.Name
		tt.Province = t.Province
		tt.City = t.City
		tt.County = t.County
		tt.Addr = t.Addr
		tt.Phone = t.Phone
		tt.Lng = t.Lng
		tt.Lat = t.Lat
		tt.FullName = t.FullName
		for _, c := range t.TUsers {
			tUser := findTUserByUserName(c.Username)
			tt.TUsers = append(tt.TUsers, tUser)
		}

		if avatar, err := os.Open("config/db/seeds/data/avatars/2.jpg"); err != nil {
			panic(fmt.Sprintf("file doesn't exist %v\n", err))
		} else {
			_ = AdminUser.Avatar.Scan(avatar)
		}

		if err := DraftDB.Create(&tt).Error; err != nil {
			log.Fatalf("create Tenant (%v) failure, got err %v", tt, err)
		}
	}
}

func createStores() {
	for _, s := range Seeds.Stores {
		store := stores.Store{}
		store.StoreName = s.Name
		store.Phone = s.Phone
		store.Email = s.Email
		store.Country = s.Country
		store.City = s.City
		store.Region = s.Region
		store.Address = s.Address
		store.Zip = s.Zip
		store.Latitude = s.Latitude
		store.Longitude = s.Longitude
		if err := DraftDB.Create(&store).Error; err != nil {
			log.Fatal(fmt.Sprintf("create store (%v) failure, got err %v", store, err))
		}
	}
}

func createMediaLibraries() {
	for _, m := range Seeds.MediaLibraries {
		medialibrary := settings.MediaLibrary{}
		medialibrary.Title = m.Title

		if file, err := openFileByURL(m.Image); err != nil {
			fmt.Printf("open file (%q) failure, got err %v", m.Image, err)
		} else {
			defer file.Close()
			_ = medialibrary.File.Scan(file)
		}

		if err := DraftDB.Create(&medialibrary).Error; err != nil {
			fmt.Println(fmt.Sprintf("create medialibrary (%v) failure, got err %v", medialibrary, err))
		}
	}
}

func createHelps() {
	helps := map[string][]string{
		"How to setup a microsite":          []string{"micro_sites"},
		"How to create a user":              []string{"users"},
		"How to create an admin user":       []string{"users"},
		"How to handle abandoned order":     []string{"abandoned_orders", "orders"},
		"How to cancel a order":             []string{"orders"},
		"How to create a order":             []string{"orders"},
		"How to upload tenant images":       []string{"tenants", "tenant_images"},
		"How to create a tenant":            []string{"tenants"},
		"How to create a discounted tenant": []string{"tenants"},
		"How to create a store":             []string{"stores"},
		"How shop setting works":            []string{"shop_settings"},
		"How to setup seo settings":         []string{"seo_settings"},
		"How to setup seo for blog":         []string{"seo_settings"},
		"How to setup seo for tenant":       []string{"seo_settings"},
		"How to setup seo for microsites":   []string{"micro_sites", "seo_settings"},
		"How to setup promotions":           []string{"promotions"},
		"How to publish a promotion":        []string{"schedules", "promotions"},
		"How to create a publish event":     []string{"schedules", "scheduled_events"},
		"How to publish a tenant":           []string{"schedules", "tenants"},
		"How to publish a microsite":        []string{"schedules", "micro_sites"},
		"How to create a scheduled data":    []string{"schedules"},
		"How to take something offline":     []string{"schedules"},
	}

	for key, value := range helps {
		helpEntry := help.QorHelpEntry{
			Title: key,
			Body:  "Content of " + key,
			Categories: help.Categories{
				Categories: value,
			},
		}
		DraftDB.Create(&helpEntry)
	}
}

func randTime() time.Time {
	num := rand.Intn(10)
	return time.Now().Add(-time.Duration(num*24) * time.Hour)
}

func openFileByURL(rawURL string) (*os.File, error) {
	if fileURL, err := url.Parse(rawURL); err != nil {
		return nil, err
	} else {
		path := fileURL.Path
		segments := strings.Split(path, "/")
		fileName := segments[len(segments)-1]

		filePath := filepath.Join(os.TempDir(), fileName)

		if _, err := os.Stat(filePath); err == nil {
			return os.Open(filePath)
		}

		file, err := os.Create(filePath)
		if err != nil {
			return file, err
		}

		check := http.Client{
			CheckRedirect: func(r *http.Request, via []*http.Request) error {
				r.URL.Opaque = r.URL.Path
				return nil
			},
		}
		resp, err := check.Get(rawURL) // add a filter to check redirect
		if err != nil {
			return file, err
		}
		defer resp.Body.Close()
		fmt.Printf("----> Downloaded %v\n", rawURL)

		_, err = io.Copy(file, resp.Body)
		if err != nil {
			return file, err
		}
		return file, nil
	}
}

func findTUserByUserName(username string) *qortenant_database.TUser {
	tUser := &qortenant_database.TUser{}
	if err := DraftDB.Where(&qortenant_database.TUser{Username: username}).First(tUser).Error; err != nil {
		log.Fatalf("can't find rabcUser with name = %q, got err %v", username, err)
	}
	return tUser
}
