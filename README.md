# Телеграм бот для получения информации про IP-адрес

---

Для запуска небходимы [Docker](https://www.docker.com/)
и [Docker-compose](https://github.com/docker/compose)

---

* [Environments](#environments)
* [Deploying](#deploying)
* [API reference](#api-reference)

---

## Environments

Создайте файл .env на основе .env.example:

```shell
cp .env.example .env
```

Содержимое .env.example:

```dotenv
DB_URI=db
DB_PORT=5432
DB_PASSWORD=myPass
DB_USER=golang
DB_DATABASE=app
DB_TIMEZONE=UTC

ADMIN_TG_ID=            # ID первого администратора

TG_BOT_TOKEN=           # Токен от @BotFather

API_PORT=8080

IPSTACK_URL=http://api.ipstack.com/
IPSTACK_ACCESS_KEY=     # API Access Key от ipstack.com

```

---

## Deploying

Для запуска достаточно выполнить:

```shell
docker-compose up
```

---

## API reference

* [/get_users](#get_users)
* [/get_user](#get_user)
* [/get_history_by_tg](#get_history_by_tg)
* [/delete_history_record](#delete_history_record)

---

### /get_users

Получение информации по всем пользователям

* **URL**

  /get_users

* **Method:**

  `GET`

* **URL Params**

  None

* **Data Params**

  None

* **Success Response:**

    * **Code:** 200 <br />
      **Content:**

    ```json
    {
        "success": true,
        "users": [
            {
                "TgID": 123456789,
                "TgUserName": "bar",
                "TgFirstName": "test1",
                "TgLastName": "",
                "TgLanguageCode": "en",
                "IsAdmin": true,
                "CreatedAt": "2020-10-04T14:23:21.446239Z",
                "UpdatedAt": "2020-10-04T14:24:13.940273Z",
                "DeletedAt": null
            },
            {
                "TgID": 987654321,
                "TgUserName": "foo",
                "TgFirstName": "test2",
                "TgLastName": "",
                "TgLanguageCode": "en",
                "IsAdmin": false,
                "CreatedAt": "2021-10-04T14:23:21.446239Z",
                "UpdatedAt": "2021-10-04T14:24:13.940273Z",
                "DeletedAt": null
            }
        ]
    }
    ```

* **Sample Call:**

  ```shell
  curl --location --request GET '127.0.0.1:8080/get_users'
  ```

---

### /get_user

Получение информации по конкретному пользователю

* **URL**

  /get_user

* **Method:**

  `GET`

* **URL Params**

  **Required:**

  `userTgID=[unsigned integer]`

* **Data Params**

  None

* **Success Response:**

    * **Code:** 200 <br />
      **Content:**

    ```json
    {
        "success": true,
        "user": {
            "TgID": 123456789,
            "TgUserName": "foo",
            "TgFirstName": "fest1",
            "TgLastName": "",
            "TgLanguageCode": "en",
            "IsAdmin": true,
            "CreatedAt": "2021-10-04T14:23:21.446239Z",
            "UpdatedAt": "2021-10-04T14:24:13.940273Z",
            "DeletedAt": null
        }
    }
    ```

* **Sample Call:**

  ```shell
  curl --location --request GET '127.0.0.1:8080/get_user?userTgID=123456789'
  ```

---



### /get_history_by_tg

Получение истории запросов конкретного пользователя

* **URL**

  /get_history_by_tg

* **Method:**

  `GET`

* **URL Params**

  **Required:**

  `userTgID=[unsigned integer]`

* **Data Params**

  None

* **Success Response:**

    * **Code:** 200 <br />
      **Content:**

      ```json
      {
          "success": true,
          "ip_check_history": [
              {
                  "ID": 1,
                  "IP": "1.2.3.4",
                  "IPInfo": {
                      "ip": "1.2.3.4",
                      "zip": "4000",
                      "city": "Brisbane",
                      "type": "ipv4",
                      "latitude": -27.467580795288086,
                      "location": {
                          "capital": "Canberra",
                          "languages": [
                              {
                                  "code": "en",
                                  "name": "English",
                                  "native": "English"
                              }
                          ],
                          "geoname_id": 2174003,
                          "calling_code": "61",
                          "country_flag": "https://assets.ipstack.com/flags/au.svg",
                          "country_flag_emoji": "🇦🇺",
                          "country_flag_emoji_unicode": "U+1F1E6 U+1F1FA"
                      },
                      "longitude": 153.02789306640625,
                      "region_code": "QLD",
                      "region_name": "Queensland",
                      "country_code": "AU",
                      "country_name": "Australia",
                      "continent_code": "OC",
                      "continent_name": "Oceania"
                  },
                  "UserTgID": 123456789,
                  "CreatedAt": "2020-10-04T15:05:30.924594Z",
                  "UpdatedAt": "2020-10-04T15:05:30.924594Z",
                  "DeletedAt": null
              }
          ]
      }
      ```

* **Sample Call:**

  ```shell
  curl --location --request GET '127.0.0.1:8080/get_history_by_tg?userTgID=123456789'
  ```

---

### /delete_history_record

Удаление записи из истории запросов

* **URL**

  /delete_history_record

* **Method:**

  `DELETE`

* **URL Params**

  **Required:**

  `ipCheckID=[unsigned integer]`

* **Data Params**

  None

* **Success Response:**

    * **Code:** 200 <br />
      **Content:**

      ```json
      {
        "success": true
      }
      ```

* **Sample Call:**

  ```shell
  curl --location --request DELETE '127.0.0.1:8080/delete_history_record?ipCheckID=2'
  ```