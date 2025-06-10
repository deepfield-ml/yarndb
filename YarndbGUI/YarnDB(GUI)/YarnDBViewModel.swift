//
//  YarnDBViewModel.swift
//  YarnDB(GUI)
//
//  Created by Gordon H on 2025/6/10.
//

import Foundation
import SwiftUI
import Yams

// Defines a record for display, with the fix for Hashable/Equatable conformance.
struct DisplayRecord: Identifiable, Hashable {
    let id: String
    let fields: [String: Any]
    var yamlString: String { try! Yams.dump(object: fields, indent: 2) ?? "YAML Error" }
    static func == (lhs: DisplayRecord, rhs: DisplayRecord) -> Bool { lhs.id == rhs.id }
    func hash(into hasher: inout Hasher) { hasher.combine(id) }
}

@MainActor
class YarnDBViewModel: ObservableObject {
    @Published var outputText: String = "Welcome to YarnDB GUI!\n"
    @Published var keyInput: String = ""
    @Published var valueInput: String = ""
    @Published var queryInput: String = "department=engineering"
    @Published var isLoading: Bool = false
    @Published var dataDirectoryPath: String = "No directory selected."
    @Published var queryResults: [DisplayRecord] = []
    @Published var selectedRecordID: String?
    
    private var executableURL: URL?
    private var dataDirectoryURL: URL?
    private let bookmarkKey = "dataDirectoryBookmark"

    init() {
        guard let url = Bundle.main.url(forResource: "yarndb", withExtension: nil) else {
            outputText = "FATAL ERROR: `yarndb` executable not found."
            return
        }
        executableURL = url
        loadDataDirectory()
    }
    
    func getExecutableURL() -> URL? { self.executableURL }
    func getDataDirectoryURL() -> URL? { self.dataDirectoryURL }

    func selectRecord(_ record: DisplayRecord?) {
        guard let record = record else { selectedRecordID = nil; return }
        self.keyInput = record.id
        self.valueInput = record.yamlString
        self.selectedRecordID = record.id
    }
    
    func resetDirectoryPermissions() {
        UserDefaults.standard.removeObject(forKey: self.bookmarkKey)
        self.queryResults.removeAll()
        self.outputText = "Permissions have been reset."
        setupDefaultDirectory()
    }

    private func loadDataDirectory() {
        guard let bookmarkData = UserDefaults.standard.data(forKey: bookmarkKey) else {
            setupDefaultDirectory()
            return
        }
        do {
            var isStale = false
            let resolvedUrl = try URL(resolvingBookmarkData: bookmarkData, options: .withSecurityScope, relativeTo: nil, bookmarkDataIsStale: &isStale)
            self.dataDirectoryURL = resolvedUrl
            self.dataDirectoryPath = resolvedUrl.path
        } catch {
            outputText = "Could not load saved directory. Please Reset Perms or Choose a new directory."
            self.dataDirectoryPath = "Error: Could not access saved path."
        }
    }
    
    private func setupDefaultDirectory() {
        do {
            let appSupportDir = try FileManager.default.url(for: .applicationSupportDirectory, in: .userDomainMask, appropriateFor: nil, create: true)
            let defaultDir = appSupportDir.appendingPathComponent("YarnDB_GUI_DefaultData")
            if !FileManager.default.fileExists(atPath: defaultDir.path) {
                try FileManager.default.createDirectory(at: defaultDir, withIntermediateDirectories: true, attributes: nil)
            }
            self.dataDirectoryURL = defaultDir
            self.dataDirectoryPath = defaultDir.path
            self.outputText += "\nUsing default data directory. Choose a custom one if you prefer."
        } catch {
            self.outputText = "FATAL ERROR: Could not create default data directory. \(error.localizedDescription)"
        }
    }

    func chooseDataDirectory() {
        let openPanel = NSOpenPanel()
        openPanel.prompt = "Choose Directory"
        openPanel.canChooseDirectories = true
        openPanel.canChooseFiles = false
        openPanel.begin { [weak self] result in
            guard let self = self, result == .OK, let url = openPanel.url else { return }
            do {
                let bookmarkData = try url.bookmarkData(options: .withSecurityScope, includingResourceValuesForKeys: nil, relativeTo: nil)
                UserDefaults.standard.set(bookmarkData, forKey: self.bookmarkKey)
                self.dataDirectoryURL = url
                self.dataDirectoryPath = url.path
                self.outputText = "✅ Set new data directory: \(url.path)"
            } catch {
                self.outputText = "❌ Error saving directory bookmark: \(error.localizedDescription)"
            }
        }
    }

    func initializeDB() { runCommand(arguments: ["init"]) }
    func getStatus() { runCommand(arguments: ["status"]) }
    func manualSave() { runCommand(arguments: ["save"]) }
    
    func setRecord() {
        guard !keyInput.isEmpty, !valueInput.isEmpty else { outputText = "Error: Key and Value required."; return }
        runCommand(arguments: ["set", keyInput, valueInput])
    }

    func getRecord() {
        guard !keyInput.isEmpty else { outputText = "Error: Key required."; return }
        runCommand(arguments: ["get", keyInput])
    }
    
    func deleteRecord() {
        guard !keyInput.isEmpty else { outputText = "Error: Key required."; return }
        runCommand(arguments: ["delete", keyInput])
    }
    
    func queryDB() {
        guard !queryInput.isEmpty else { outputText = "Error: Query required."; return }
        queryResults.removeAll()
        runCommand(arguments: ["query", queryInput])
    }
    
    func createIndex() {
        guard !queryInput.contains("=") else { outputText = "Error: Provide key name only for index."; return }
        guard !queryInput.isEmpty else { outputText = "Error: Key name required."; return }
        runCommand(arguments: ["index", queryInput])
    }

    private func runCommand(arguments: [String]) {
        guard let executableURL = executableURL, let dataDirectoryURL = dataDirectoryURL else {
            outputText = "Error: Database paths not configured."
            return
        }
        
        isLoading = true
        outputText = "Running: yarndb \(arguments.joined(separator: " "))..."

        guard dataDirectoryURL.startAccessingSecurityScopedResource() else {
            outputText = "Error: Could not gain access to the data directory. Please choose it again or reset perms."
            isLoading = false
            return
        }
        defer { dataDirectoryURL.stopAccessingSecurityScopedResource() }

        Task.detached(priority: .userInitiated) {
            let process = Process()
            process.executableURL = executableURL
            process.arguments = arguments
            process.currentDirectoryURL = dataDirectoryURL
            
            let outputPipe = Pipe()
            let errorPipe = Pipe()
            process.standardOutput = outputPipe
            process.standardError = errorPipe

            do {
                try process.run()
                process.waitUntilExit()
                let outputData = outputPipe.fileHandleForReading.readDataToEndOfFile()
                let errorData = errorPipe.fileHandleForReading.readDataToEndOfFile()
                let outputString = String(data: outputData, encoding: .utf8) ?? ""
                let errorOutput = String(data: errorData, encoding: .utf8) ?? ""
                
                let combinedOutput = """
                [STDOUT]:
                \(outputString.isEmpty ? "<No output>" : outputString)
                [STDERR]:
                \(errorOutput.isEmpty ? "<No errors>" : errorOutput)
                """

                var parsedRecords: [DisplayRecord] = []
                if arguments.first == "query", !outputString.isEmpty {
                    let docs = try Yams.load_all(yaml: outputString)
                    for doc in docs {
                        if let dict = doc as? [String: Any], let key = dict.keys.first, let fields = dict[key] as? [String: Any] {
                            parsedRecords.append(DisplayRecord(id: key, fields: fields))
                        }
                    }
                }
                
                await MainActor.run {
                    self.outputText = combinedOutput.trimmingCharacters(in: .whitespacesAndNewlines)
                    if !parsedRecords.isEmpty { self.queryResults = parsedRecords }
                    self.isLoading = false
                }
            } catch {
                await MainActor.run {
                    self.outputText = "Failed to run command: \(error.localizedDescription)"
                    self.isLoading = false
                }
            }
        }
    }
}
