// MongoDB Initialization Script

// Switch to admin database
db = db.getSiblingDB("admin");

// Create application database
var appDb = db.getSiblingDB("droadmap_app");

// Create tenants collection with schema validation
appDb.createCollection("tenants", {
    validator: {
        $jsonSchema: {
            bsonType: "object",
            required: ["name", "created_at"],
            properties: {
                _id: { bsonType: "objectId" },
                name: { bsonType: "string", description: "Tenant name" },
                metadata: { bsonType: "object" },
                created_at: { bsonType: "date" },
                updated_at: { bsonType: "date" }
            }
        }
    }
});

// Create documents collection with schema validation
appDb.createCollection("documents", {
    validator: {
        $jsonSchema: {
            bsonType: "object",
            required: ["tenant_name", "file_name", "uploaded_at"],
            properties: {
                _id: { bsonType: "objectId" },
                tenant_name: { bsonType: "string" },
                file_name: { bsonType: "string" },
                file_size: { bsonType: "int" },
                extracted_text: { bsonType: "string" },
                summary: { bsonType: "string" },
                storage_path: { bsonType: "string" },
                storage_url: { bsonType: "string" },
                uploaded_at: { bsonType: "date" },
                is_deleted: { bsonType: "bool" },
                deleted_at: { bsonType: ["date", "null"] }
            }
        }
    }
});

// Create indexes for performance
appDb.tenants.createIndex({ name: 1 });
appDb.tenants.createIndex({ created_at: -1 });

appDb.documents.createIndex({ tenant_name: 1 });
appDb.documents.createIndex({ uploaded_at: -1 });
appDb.documents.createIndex({ file_name: 1 });
appDb.documents.createIndex({ is_deleted: 1 });

print("✅ MongoDB initialization complete");
