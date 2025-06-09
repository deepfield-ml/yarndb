# YarnDB: A Thread-Safe, High-Speed In-Memory YAML Database
[![Ask DeepWiki](https://deepwiki.com/badge.svg)](https://deepwiki.com/deepfield-ml/yarndb)
`YarnDB` is a production-ready, in-memory database that stores data in YAML files. It supports generic YAML structures, concurrent reading, thread-safe operations, indexing, transactions, and a powerful CLI.

## Features
- **Generic YAML Support**: Stores any valid YAML data without predefined schemas.
- **High-Speed Reading**: Concurrently loads YAML files with one goroutine per file.
- **Thread Safety**: Uses `sync.RWMutex` and `sync.Mutex` for safe concurrent access.
- **In-Memory Processing**: All operations (CRUD, queries) run in RAM.
- **Indexing**: Creates indexes on keys for fast queries.
- **Transactions**: Supports atomic updates with commit/rollback.
- **Auto-Save**: Saves to disk periodically (configurable).
- **CLI**: Intuitive commands for database operations using `cobra`.
## Home Brew Installation and Usage
Install via Homebrew:

```bash
brew install yarndb
```

Initialize :

```bash
yarndb init
```
（Creates and loads in a ` data ` folder in current directory)
Add a record:

```bash
yarndb set record1_1 "name: Alice Smith\ndepartment: engineering\nage: 30"
```

Query records:

```bash
yarndb query department=engineering
```

Create an index:

```bash
yarndb index department
```

Run a transaction:

```bash
yarndb trans
tx> set record1_2 "name: Bob Jones\ndepartment: marketing"
tx> commit
```

Check database status:

```bash
yarndb status
```

## Manual Installation
1. Ensure Go 1.20+ is installed.
2. Create the project directory:
   ```bash
   mkdir yarndb
   cd yarndb
   ```
3. Save the provided files: `main.go`, `datastore.go`, `save.go`, `reader.go`, `writer.go`, `go.mod`, `config.yaml`, `data/records_1.yaml`, `data/records_2.yaml`, `README.md`.
4. Initialize the module:
   ```bash
   go mod tidy
   ```

## Usage for Manual Installation
Build the application with:
```bash
go build -o yarndb
```

### Commands
- **Initialize YarnDB**:
  ```bash
  ./yarndb init
  ```
- **Set a record** (create/update):
  ```bash
  ./yarndb set record1_3 "name: Charlie Green\ndepartment: engineering\nage: 35"
  ```
- **Get a record**:
  ```bash
  ./yarndb get record1_1
  ```
- **Delete a record**:
  ```bash
  ./yarndb delete record1_2
  ```
- **Query records** (e.g., by department):
  ```bash
  ./yarndb query department=engineering
  ```
- **Create an index** (for faster queries):
  ```bash
  ./yarndb index department
  ```
- **Start a transaction**:
  ```bash
  ./yarndb trans
  tx> set record1_4 "name: Dave Black\ndepartment: marketing"
  tx> delete record1_2
  tx> commit
  ```
- **Manually save**:
  ```bash
  ./yarndb save
  ```
- **Check status**:
  ```bash
  ./yarndb status
  ```

### Configuration
Edit `config.yaml` to customize:
```yaml
data_dir: data
auto_save_interval: 60
log_level: info
```
Override with flags:
```bash
./yarndb --data-dir custom_data --auto-save-interval 30 --log-level debug init
```

## Example Workflow
1. Initialize YarnDB:
   ```bash
   ./yarndb init
   ```
2. Add a record:
   ```bash
   ./yarndb set record1_5 "name: Eve Blue\ndepartment: engineering\nskills: [Go, Python]"
   ```
3. Create an index on `department`:
   ```bash
   ./yarndb index department
   ```
4. Query engineers:
   ```bash
   ./yarndb query department=engineering
   ```
5. Start a transaction to update multiple records:
   ```bash
   ./yarndb trans
   tx> set record1_6 "name: Frank Red\ndepartment: marketing"
   tx> delete record1_5
   tx> commit
   ```
6. Check status:
   ```bash
   ./yarndb status
   ```

## Notes
- **Performance**: Optimized for high-speed reading with concurrent file loading and in-memory indexing.
- **Thread Safety**: Safe for concurrent reads and writes.
- **Limitations**: Best for small to medium datasets. For large-scale production, consider SQLite or Redis.
- **Logging**: Logs are written to `yarndb.log` and console.

## Requirements: 
- ** At Least 8GB RAM Suggested
- ** Intel Cored-Mac (Only Intel-Cored Supported Currently)
## License
Apache 2.0

Produced by Deepfield ML, Gordon.H and Will.C
