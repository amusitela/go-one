package model

//执行数据迁移

func migration() {
    _ = DB.AutoMigrate(&User{})
    _ = DB.AutoMigrate(&RefreshToken{})
}
