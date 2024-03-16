export interface LoginParams {
  phone_number?: string
  password?: string
}

export interface UserInfo {
  access_token: string
  token_type: string
  refresh_token: string
}

export interface PasswordParams {
  old_password?: string
  new_password?: string
}
