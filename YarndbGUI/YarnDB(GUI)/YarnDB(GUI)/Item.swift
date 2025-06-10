//
//  Item.swift
//  YarnDB(GUI)
//
//  Created by Gordon H on 2025/6/10.
//

import Foundation
import SwiftData

@Model
final class Item {
    var timestamp: Date
    
    init(timestamp: Date) {
        self.timestamp = timestamp
    }
}
