use serde::{Serialize, Deserialize};
use mongodb::{Client, options::{ClientOptions, IndexModel, UpdateOptions}};
use std::collections::HashMap;
use std::error::Error;
use std::sync::{Arc, RwLock};
use std::marker::Sync;

const FLAGS_COLLECTION: &'static str = "flags";

#[derive(Debug)]
pub enum Type {
    Mem,
    Mongo
}

impl std::str::FromStr for Type {
    type Err = &'static str;
    fn from_str(s: &str) -> Result<Self, Self::Err> {
        match s {
            "mem" => Ok(Type::Mem),
            "mongo" => Ok(Type::Mongo),
            _ => Err("Error parsing storage type"),
        }
    }
}

#[derive(Hash, PartialEq, Eq)]
struct FlagKey {
    name: String,
    service_name: String,
}

#[derive(Serialize, Deserialize, Debug, Clone)]
pub struct Flag {
    name: String,
    #[serde(rename="serviceName")]
    service_name: String,
    #[serde(rename="rawValue")]
    raw_value: String,
    value: bool,
}

pub trait Store: Clone + Sync {
    fn get_flags(&self, service_name: Option<String>) -> Result<Vec<Flag>, Box<dyn Error>>;
    fn store(&mut self, flags: Vec<Flag>, initial: bool) -> Result<(), Box<dyn Error>>;
}

#[derive(Clone)]
pub struct MemStore {
    data: Arc<RwLock<HashMap<FlagKey, Flag>>>,
}

impl Store for MemStore {
    fn get_flags(&self, service_name: Option<String>) -> Result<Vec<Flag>, Box<dyn Error>> {
        let data = self.data.read().unwrap();
        let mut values = Vec::with_capacity(data.len());
        let no_service_name = service_name.is_none();
        let service_name = service_name.unwrap_or(String::from(""));
        for val in data.values() {
            if no_service_name || val.service_name == service_name || val.service_name == ""  {
                values.push(val.clone());
            }
        }
        Ok(values)
    }

    fn store(&mut self, flags: Vec<Flag>, initial: bool) -> Result<(), Box<dyn Error>> {
        let mut data = self.data.write().unwrap();
        for f in flags {
            let val = f.clone();
            let key = FlagKey{name: f.name, service_name: f.service_name};
            if !initial || !data.contains_key(&key) {
                data.insert(key, val);
            }
        }
        Ok(())
    }
}

impl MemStore {
    pub fn new() -> impl Store {
        MemStore{data: Arc::new(RwLock::new(HashMap::new()))}
    }
}

#[derive(Clone)]
pub struct MongoStore {
    db: String,
    client: Client,
}

impl Store for MongoStore {
    fn get_flags(&self, service_name: Option<String>) -> Result<Vec<Flag>, Box<dyn Error>> {
        let db = self.client.database(&self.db);
        let coll = db.collection(FLAGS_COLLECTION);
        let filter = match service_name {
            Some(service_name) => doc! {"serviceName": {"$or": [{"$eq": service_name}, {"$eq": ""}] }},
            None => doc! {},
        };

        let cursor = coll.find(filter, None)?;
        let mut ret = Vec::new();

        for res in cursor {
            let f = bson::from_bson(bson::Bson::Document(res?))?;
            ret.push(f);
        }
        Ok(ret)
    }

    fn store(&mut self, flags: Vec<Flag>, initial: bool) -> Result<(), Box<dyn Error>> {
        let db = self.client.database(&self.db);
        let coll = db.collection(FLAGS_COLLECTION);

        for f in flags {
            coll.update_one(doc! {}, doc! {"$set": bson::to_bson(&f)?}, Some(UpdateOptions::builder().upsert(true).build()))?;
        }

        Ok(())
    }
}

impl MongoStore {
    pub fn new(url: &str) -> Result<impl Store, Box<dyn Error>> {
        let dbopts = ClientOptions::parse(&url)?;
        let db = dbopts.clone().credential.map(|c| c.source).flatten().unwrap_or(String::from("featureToggles"));
        let client = Client::with_options(dbopts)?;

        client.database(&db).collection(FLAGS_COLLECTION).create_indexes(vec![IndexModel{keys: doc! {"name": 1, "service_name": 1}, options: Some(doc! {"unique": true})}])?;

        Ok(MongoStore{db, client})
    }
}