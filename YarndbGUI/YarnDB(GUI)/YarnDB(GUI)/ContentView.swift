//
//  ContentView.swift
//  YarnDB(GUI)
//
//  Created by Gordon H on 2025/6/10.
//

import SwiftUI

struct ContentView: View {
    @StateObject private var viewModel = YarnDBViewModel()
    @Environment(\.openWindow) var openWindow
    @State private var resetPermsTrigger = false

    var body: some View {
        HSplitView {
            // Left Panel: Controls
            VStack(alignment: .leading, spacing: 15) {
                Text("YarnDB Control Panel")
                    .font(.title).fontWeight(.bold)

                GroupBox("Database") {
                    VStack(alignment: .leading) {
                        Text("Data Directory:").font(.headline)
                        Text(viewModel.dataDirectoryPath).font(.caption.monospaced()).padding(5).background(Color.primary.opacity(0.1)).cornerRadius(4).lineLimit(2).help(viewModel.dataDirectoryPath)
                        HStack {
                            Button("Initialize", systemImage: "power.circle", action: { viewModel.initializeDB() })
                            Button("Status", systemImage: "chart.bar.doc.horizontal", action: { viewModel.getStatus() })
                            Button("Save", systemImage: "square.and.arrow.down", action: { viewModel.manualSave() })
                            Spacer()
                            Button("Reset Perms", systemImage: "exclamationmark.lock", action: { self.resetPermsTrigger.toggle() })
                                .help("Reset saved folder permissions and revert to the default directory.").tint(.orange)
                            Button("Choose...", action: { viewModel.chooseDataDirectory() })
                        }
                    }
                    .buttonStyle(.bordered)
                }

                // --- UPDATED BUTTON TO LAUNCH THE TERMINAL ---
                GroupBox("Tools") {
                    HStack {
                        Text("Open a terminal in the data directory.")
                            .font(.caption)
                            .foregroundColor(.secondary)
                        Spacer()
                        Button("Open Terminal...", systemImage: "terminal.fill") {
                            if let exeURL = viewModel.getExecutableURL(), let dataURL = viewModel.getDataDirectoryURL() {
                                let config = TerminalWindowConfig(executableURL: exeURL, dataDirectoryURL: dataURL)
                                openWindow(id: "terminal-window", value: config)
                            }
                        }
                        .tint(.green)
                    }
                }
                // --- END OF UPDATE ---
                
                GroupBox("Record Operations (C.R.U.D)") {
                    VStack(alignment: .leading) {
                        TextField("Key (e.g., record1_1)", text: $viewModel.keyInput)
                        Text("Value (YAML Format):").font(.caption).foregroundColor(.secondary)
                        TextEditor(text: $viewModel.valueInput).frame(height: 100).font(.system(.body, design: .monospaced)).border(Color.gray.opacity(0.3), width: 1).cornerRadius(4)
                        HStack {
                            Button("Set", systemImage: "square.and.pencil", action: { viewModel.setRecord() })
                            Button("Get", systemImage: "arrow.down.doc", action: { viewModel.getRecord() })
                            Button("Delete", systemImage: "trash", action: { viewModel.deleteRecord() }).tint(.red)
                        }
                        .buttonStyle(.bordered)
                    }
                }
                
                GroupBox("Query & Index") {
                    VStack(alignment: .leading) {
                        TextField("Key or Key=Value (e.g., department=eng)", text: $viewModel.queryInput)
                        HStack {
                            Button("Query", systemImage: "magnifyingglass", action: { viewModel.queryDB() })
                            Button("Create Index", systemImage: "key.horizontal", action: { viewModel.createIndex() })
                        }
                        .buttonStyle(.bordered)
                    }
                }
                
                Spacer()
            }
            .padding().frame(minWidth: 350, maxWidth: 450).disabled(viewModel.isLoading)

            // Right Panel: Results and Output (No changes here)
            VStack(alignment: .leading) {
                // ... (This section is unchanged) ...
                Text("Query Results (\(viewModel.queryResults.count))").font(.headline)
                List(selection: $viewModel.selectedRecordID) { /* ... */ }
                Divider().padding(.vertical, 5)
                HStack { /* ... */ }
                TextEditor(text: .constant(viewModel.outputText))
            }
            .padding().frame(minWidth: 400)
        }
        .frame(minWidth: 800, minHeight: 550)
        .onChange(of: resetPermsTrigger) { _ in viewModel.resetDirectoryPermissions() }
    }
}
