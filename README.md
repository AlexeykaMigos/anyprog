# Цель задания:
Разработать RESTful-сервис на языке Go с использованием PostgreSQL, который позволит управлять товарами: просматривать список товаров, добавлять новые товары, обновлять данные о товарах и выполнять откат до предыдущих версий товаров. Авторизация и аутентификация не требуются.


MAKEFILE: 

Чтобы сгенерировать .env пишем `make .env`.

Если у вас есть .env, удалить можно написав `make clean-env`.

`make build` Чтобы сгенерировать сервис.

`make up` Чтобы запустить сервис.

`make down` Чтобы выключить сервис.

