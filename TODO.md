# TODO:

## Docker (Docker-compose)
- [x] add go service
- [x] add PostgreSQL service
- [x] add network

## Tg bot
- [x] add user create/update on new message
- [x] add safe sending
- [x] pretty print IP-addr check
- [x] user menu
  - [x] create check IP-addr command/btn
  - [x] create list of uniq IP-addr checks command/btn
  - [x] create list of results uniq IP-addr checks command/btn
  
- [x] admin menu
  - [x] create broadcast msg btn
  - [x] create add/remove admins btn
  - [x] create list of uniq IP-addr checks btn
  
## Backend
- [x] API
  - [x] create /get_users - return full info about all users
  - [x] create /get_user?id=1 - return info about user with id=1
  - [x] create /get_history_by_tg?tg_id=123 - return request history and checks results of user by tg_id
  - [x] create /delete_history_record? - delete specific records from history
  - [x] add middleware 4 check query parameters

## DB
- [x] create schema
- [x] create migrations
- [x] create methods for User
- [x] create methods for IPCheck

## IP check
- [x] create getIPInfo func

## Logging
- [x] add logging errors in DB