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
    make compine_consumer
   ./consumer_svc
    ```
   
# dev env
- macos High Sierra
- golang 1.13.4
- docker desktop v2.1.0.5 (40693)
    - Docker Engine: 19.03.5
    - Docker Compose: 1.24.1