# Golang Market Data Fetcher
# Project Overview
This project is a market data fetcher developed in Golang, tailored for integrating with various cryptocurrency exchanges. Although the current AMM system has not yet directly utilized this repository for its functionality, the program has already demonstrated the capability to connect to Binance exchange's real-time market data and exhibits excellent extensibility for seamless adaptation to other exchanges.

# Core Features
* WebSocket Real-Time Data Integration: Establishes a WebSocket (ws) connection with the Binance exchange to retrieve market data in real time.

* Trading Pair Subscription Management: Reads symbol (trading pair) information from the database, automatically performing relevant subscription operations to ensure accurate and timely data updates.

* Exception Handling: Effectively deals with exceptions such as subscription timeouts, ensuring the program runs smoothly.

* Integration with Otmoic Application: Utilizes Redis channels to communicate with the Otmoic application, triggering actions upon token or configuration changes.

# Runtime Environment
Like other AMM programs, this program relies on a set of predefined environment variables within an OS Pod during runtime:

```Bash
STATUS_KEY=obridge-amm-market-status-report-price
OBRIDGE_LPNODE_DB_REDIS_MASTER_SERVICE_HOST=******
REDIS_PORT=6379
REDIS_PASSWORD=******
SERVICE_PORT=18080
MONGODB_HOST=******
MONGODB_PORT=27017
MONGODB_ACCOUNT=******
MONGODB_PASS=******
MONGODB_DBNAME_LP_STORE=******
```
>Please replace the placeholder password (******) with the actual value to ensure system security.
