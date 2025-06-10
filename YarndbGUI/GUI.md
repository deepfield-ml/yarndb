
# YarnDB GUI for macOS



A native macOS SwiftUI interface for the **[YarnDB](https://github.com/deepfield-ml/yarndb)** in-memory database.

This application provides a user-friendly graphical interface that wraps the `yarndb` command-line tool. It allows you to manage your database—set records, run queries, create indexes, and check status—all without ever touching the terminal. The app is fully self-contained and does not require a separate installation of YarnDB.

## Features

*   **Visual Interface:** Manage all major YarnDB commands through an intuitive point-and-click interface.
*   **Self-Contained:** The `yarndb` executable is bundled inside the app. No need to install Homebrew or Go.
*   **Custom Data Directory:** Choose any folder on your Mac to store your database files. The app securely remembers your choice between launches.
*   **Real-time Command Output:** See the standard output and error messages from every command you run.

## Getting Started: First Launch

Follow these steps to set up your database for the first time.

#### 1. Choose a Data Directory

When you first launch the app, it will use a default, hidden folder. It's highly recommended to create and select your own dedicated folder for your database.

1.  Create a new folder somewhere convenient (e.g., on your Desktop) and name it `MyYarnDB`.
2.  In the app, click the **"Choose..."** button in the "Database" panel.
3.  A file dialog will open. Navigate to and select the `MyYarnDB` folder you just created.

The "Data Directory" path in the app will update to your new folder. The app will remember this location every time you open it.

#### 2. Initialize the Database

Now that you have a data directory, you need to initialize it.

1.  Click the **"Initialize"** button.
2.  The `yarndb` command will create a `data` subfolder inside your chosen directory (`MyYarnDB/data/`). This is where your YAML record files will be stored.
3.  The output panel will show a confirmation message.

You are now ready to use the database!

## How to Use the App

The application is divided into a control panel on the left and an output panel on the right.

### Database Panel

This panel manages the database location and its overall state.

*   **Data Directory:** Shows the currently active folder for your database.
*   **Choose...:** Opens a dialog to select a new data directory.
*   **Initialize:** Sets up the necessary `data` folder in the current directory. Use this when starting in a new directory.
*   **Status:** Shows database statistics, such as the total number of records and a list of existing indexes.

### Record Operations (CRUD)

This panel is for Creating, Reading, Updating, and Deleting individual records.

*   **Key:** The unique identifier for a record (e.g., `record_1`, `user_alice`).
*   **Value (YAML Format):** The content of your record, written in valid YAML. Each key-value pair must be on a new line.
    *   **Example:**
        ```yaml
        name: Alice Smith
        department: engineering
        age: 30
        skills: [Go, Python]
        ```
*   **Set:** Creates a new record with the given Key and Value. If the Key already exists, it will be updated.
*   **Get:** Retrieves and displays the content of the record matching the Key.
*   **Delete:** Removes the record matching the Key from the database.

### Query & Index Panel

This panel is for searching records and optimizing search performance.

*   **Input Field:** This field serves two purposes:
    1.  **For Querying:** Enter a `key=value` pair (e.g., `department=engineering`).
    2.  **For Indexing:** Enter only the `key` you want to index (e.g., `department`).
*   **Query:** Searches all records and returns the ones that match the `key=value` criteria. For faster queries, create an index first.
*   **Create Index:** Creates an index on the specified key. This dramatically speeds up future queries on that key. You only need to do this once per key.

### Output Panel

This panel on the right shows the direct output from the `yarndb` command-line tool.

*   **`[STDOUT]` (Standard Output):** Shows the normal results of a command, like the content of a record or a list of query results.
*   **`[STDERR]` (Standard Error):** Shows any error messages. If a command fails, check here to see why.
*   A spinning progress indicator will appear next to the "Output" title while a command is running.

## Troubleshooting

*   **"Permission Denied" or Access Errors:** If you see an error related to file access in the output panel, your app may have lost permission to the data directory (this can happen after system updates). Simply click the **"Choose..."** button and re-select your data directory to grant permission again.
*   **App Fails to Start:** The application bundle may be corrupted, meaning the `yarndb` executable inside it is missing or damaged. Please re-download or rebuild the application.
