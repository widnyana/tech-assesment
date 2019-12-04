# Kumparan Skill Test


# How to
1. run the containers
    ```bash
    make containers
    ```

2. run api service
    ```bash
    make compile_api
    ./api_svc
    ```

3. run queue consumer
    ```bash
    make compile_consumer
   ./consumer_svc
    ```
   
4. run feeder

    in case you want to insert bulk data 
    ```bash
       go run feeder/main.go 
    ```

5. test the pagination:
    ```bash
    curl "localhost:3000/news?page=1" -s | python -m json.tool 
    ```
# dev env
- macos High Sierra
- golang 1.13.4
- docker desktop v2.1.0.5 (40693)
    - Docker Engine: 19.03.5
    - Docker Compose: 1.24.1