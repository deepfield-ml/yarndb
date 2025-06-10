//
//  TerminalView.swift
//  YarnDB(GUI)
//
//  Created by Gordon H on 2025/6/10.
//

import SwiftUI

struct TerminalView: View {
    @StateObject var viewModel: TerminalViewModel = TerminalViewModel()
    @Environment(\.dismiss) var dismiss
    
    // We still need these to construct the preset commands.
    let executableURL: URL
    let dataDirectoryURL: URL

    var body: some View {
        VStack(spacing: 0) {
            ScrollViewReader { proxy in
                ScrollView {
                    Text(viewModel.terminalOutput)
                        .font(.system(.body, design: .monospaced))
                        .frame(maxWidth: .infinity, alignment: .leading)
                        .padding()
                        .id("bottom")
                }
                .onChange(of: viewModel.terminalOutput) { _ in
                    proxy.scrollTo("bottom", anchor: .bottom)
                }
            }

            // --- UPDATED PRESET COMMANDS ---
            HStack(spacing: 4) {
                Text("YarnDB Presets:").font(.caption).foregroundColor(.secondary)
                Button("set...") {
                    viewModel.commandInput = "'\(executableURL.path)' set <key> \"name: value\""
                }
                Button("query...") {
                    viewModel.commandInput = "'\(executableURL.path)' query <field=value>"
                }
                Button("status") {
                    viewModel.commandInput = "'\(executableURL.path)' status"
                }
                // This preset now just inserts the command to start a transaction.
                Button("trans...") {
                    viewModel.commandInput = "'\(executableURL.path)' trans"
                }
                Spacer()
            }
            .buttonStyle(.plain)
            .padding(.horizontal)
            .padding(.top, 8)

            // --- UPDATED INPUT AREA ---
            HStack(spacing: 0) {
                // The prompt is now handled by the shell's environment (PS1)
                TextField("Enter command and press Enter", text: $viewModel.commandInput)
                    .font(.system(.body, design: .monospaced))
                    .textFieldStyle(.plain)
                    .padding(.leading)
                    .onSubmit(viewModel.sendCommand)
                
                Button("Close", action: { dismiss() }).padding(.trailing)
            }
            .padding(.vertical, 10)
            .background(.regularMaterial)
        }
        .frame(minWidth: 600, idealWidth: 700, minHeight: 400, idealHeight: 500)
        .onAppear {
            viewModel.beginSession(dataDirectoryURL: dataDirectoryURL)
        }
        .onDisappear {
            viewModel.endSession()
        }
        .disabled(!viewModel.isSessionActive)
    }
}
