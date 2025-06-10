//
//  YarnDB_GUIApp.swift
//  YarnDB(GUI)
//
//  Created by Gordon H on 2025/6/10.
//
import SwiftUI

// Renamed to reflect its new, general-purpose role.
struct TerminalWindowConfig: Hashable, Codable {
    let executableURL: URL
    let dataDirectoryURL: URL
}

@main
struct YarnDB_GUIApp: App {
    var body: some Scene {
        WindowGroup {
            ContentView()
        }

        // Renamed the window title and ID. It now opens our new TerminalView.
        WindowGroup("Terminal", id: "terminal-window", for: TerminalWindowConfig.self) { $config in
            if let config = config {
                TerminalView(
                    // Pass the config to the new TerminalView
                    executableURL: config.executableURL,
                    dataDirectoryURL: config.dataDirectoryURL
                )
            }
        }
    }
}
