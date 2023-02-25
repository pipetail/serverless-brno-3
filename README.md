# Serverless Brno #3 demo project

Tenhle projekt simuluje jednoduché Websocket api, které umí komunikovat
dvěmi směry.

## Databáze spojení
Spojení pro vás drží samotná API Gateway a Lambda funkce jsou
spouštěny pouze, když směrem od klienta projde nějaká zpráva,
která odpovídá nějakému pravidlu.

Pokud ovšem chcete komunikovat druhým směrem, tak potřebujete znát
identifikátor spojení, kde lze daného uživatele a nebo zařízení zastihnout.

V `$connect` handleru tedy dochází k vytváření záznamů v jednoduché
DynamoDB tabulce, v které můžeme jednoduše zjitit, komu jaké spojení patří. Následné odeslání zprávy je poté realizováno pomocí API Gateway management API, kde zkrátka do daného spojení pošleme nějaký payload.

Pokud dojde k ukončení spojení, tak je zavolán `$disconnect` handler
a ten záznam v DynamoDB smaže. Všechna využití databáze mezi `$connect` a `$disconnect` jsou pouze na vás. V tomto repozitáři
simulujeme scénář, kdy chceme prohlížeči používaném určitým
uživetelem říct, že si může z REST api něco stáhnout.

## Autentizace a autorizace
Pro ověření uživatele můžete zvážit dvě metody:

1. ověření query stringu během `$connect` fáze a ukončení chybou, pokud uživatel neposkytl správné informace. Tuto kontrolu lze implementovat authorizerem a nebou
samotnou connect funkcí. Tento mechanismus lze použít pouze na `$connect`, jelikož
po upgrade spojení už nic jako query stringy či hlavičky neexistuje. Tohle je princip
platný pro všechna websocket řešení. Při implementaci je taktéž nutné
myslet na to, co podporují prohlížeče. Websocket API v prohlížečích
totiž neumožňuje nastavení hlaviček, zatímco třeba `websocat` to zvládne.

2. vlastním protokolem, v zásadě vytvoříte speciální akci `authorize`, která označí
(a nebo neoznačí) aktuální spojení za autorizované. Tato metoda ovšem vyžaduje přístup
k perzistentní vrstvě při zpracování všech příchozích zpráv.

## Směry komunikace

### Zprávy zaslané uživatelem

Prvním směrem jsou zprávy od uživatele, který otevřel spojení. API Gateway
rozumí jsonu a podle určeného klíče dokáže posílat zprávy na definované
integrace. V tomto konkrétním případě můžete poslat zprávu

```json
{"action": "ping", "name": "Stepan"}
```

a vrátí se vám `pong` (nebo něco takového).

Pokud pošlete cokoliv jiného, tak se zpráva předá do integrace nastavené
pro route `$default` a dojde ke vrácení zprávy, že daný kus API
nebyl implementován.

Pokud potřebujete v rámci API řešit více aktivit, tak stačí přidat route
s názvem aktivity do API Gateway a naimplementovat kód obsluhující tuto
aktivitu.

### Zprávy zaslané aplikací

Websockety samozřejmě fungují i druhou stranou, takže pro demostraci tohoto
směru zde máme 2 SQS fronty, které pošlou zprávu

1. uživateli do všech jeho spojení (`NotifyUser`)


    ```json
    {"userId": "1234", "data": "cokoliv"}
    ```

2. do daného spojení (`NotifyConnection`)

    ```json
    {"connectionId": "f97yeeMZDoECGqg=", "data": "cokoliv"}
    ```

Druhá fronta je zároveň interně používaná ostatními funkcemi, čili pokud chcete
notifikovat uživatele, tak odešlete zprávu do `NotifyUser` a funkce odpovědná
za tuto aktivitu najde v databázi všechna spojení pro daného uživatele a
odešle odpovídající počet zpráv do fronty `NotifyConnection`.

## deployment

Pro nasazení tohoto stacku potřebujete jen `sam-cli` a nějaký AWS account.
Samotné nasazení je pak jen o

```bash
sam build
sam deploy --guided
```

Jako výstup `deploy` příkazu dostanete `wss://...` adresu API Gateway
endpointu, na který se můžete připojit třeba programem `websocat`.