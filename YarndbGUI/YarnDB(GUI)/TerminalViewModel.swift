//
//  TerminalViewModel.swift
//  YarnDB(GUI)
//
//  Created by Gordon H on 2025/6/10.
//

import Foundation
import Combine

@MainActor
class TerminalViewModel: ObservableObject {
    @Published var terminalOutput: String = ""
    @Published var isSessionActive: Bool = false
    @Published var commandInput: String = ""

    private var shellProcess: Process?
    private var inputPipe = Pipe()
    private var outputPipe = Pipe()
    private var outputCancellable: AnyCancellable?

    // The logic here is mostly the same, but we launch a shell.
    func beginSession(dataDirectoryURL: URL) {
        guard !isSessionActive else { return }

        isSessionActive = true
        
        guard dataDirectoryURL.startAccessingSecurityScopedResource() else {
            terminalOutput = "FATAL: Could not gain access to the data directory.\n"
            isSessionActive = false
            return
        }

        shellProcess = Process()
        
        // --- THE KEY CHANGE: LAUNCH A REAL SHELL ---
        shellProcess?.executableURL = URL(fileURLWithPath: "/bin/zsh")
        // "-il" starts an interactive login shell, which loads the user's profile (.zshrc)
        shellProcess?.arguments = ["-il"]
        // --- END OF KEY CHANGE ---
        
        shellProcess?.currentDirectoryURL = dataDirectoryURL
        shellProcess?.standardInput = inputPipe
        shellProcess?.standardOutput = outputPipe
        shellProcess?.standardError = outputPipe

        // Set up the environment to have a clean prompt and identify the shell
        var environment = ProcessInfo.processInfo.environment
        environment["PS1"] = "[YarnDB] $ " // Set a custom, clear prompt
        environment["TERM"] = "xterm-256color"
        shellProcess?.environment = environment

        outputCancellable = NotificationCenter.default.publisher(for: .NSFileHandleDataAvailable, object: outputPipe.fileHandleForReading)
            .sink { [weak self] _ in
                let data = self?.outputPipe.fileHandleForReading.availableData
                if let output = String(data: data ?? Data(), encoding: .utf8), !output.isEmpty {
                    DispatchQueue.main.async { self?.terminalOutput += output }
                }
                self?.outputPipe.fileHandleForReading.waitForDataInBackgroundAndNotify()
            }

        outputPipe.fileHandleForReading.waitForDataInBackgroundAndNotify()

        do {
            try shellProcess?.run()
            terminalOutput = "Shell session started in: \(dataDirectoryURL.path)\n"
        } catch {
            terminalOutput += "ERROR starting shell: \(error.localizedDescription)\n"
            isSessionActive = false
        }
        
        shellProcess?.terminationHandler = { [weak self] _ in
            DispatchQueue.main.async {
                self?.endSession(gracefully: true)
                dataDirectoryURL.stopAccessingSecurityScopedResource()
            }
        }
    }

    func sendCommand() {
        guard isSessionActive, !commandInput.isEmpty else { return }
        let commandWithNewline = commandInput + "\n"
        guard let data = commandWithNewline.data(using: .utf8) else { return }
        
        do {
            try inputPipe.fileHandleForWriting.write(contentsOf: data)
            commandInput = ""
        } catch {
            terminalOutput += "\nERROR writing to shell: \(error.localizedDescription)\n"
        }
    }

    func endSession(gracefully: Bool = false) {
        if !gracefully, shellProcess?.isRunning == true {
            terminalOutput += "\nShell session terminated by user.\n"
            shellProcess?.terminate()
        } else {
            terminalOutput += "\nShell session ended.\n"
        }
        isSessionActive = false
        outputCancellable?.cancel()
        shellProcess = nil
    }
}
