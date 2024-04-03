use std::{collections::HashMap, fmt::Debug};

use serde::Deserialize;

use crate::models::{Job, JobResult};

pub type JobName = String;

pub struct Worker {
    handlers: HashMap<JobName, Box<dyn Fn(Job<rmpv::Value>) -> JobResult>>,
}

impl Debug for Worker {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        f.debug_struct("Worker")
            .field("handlers", &self.handlers.keys().collect::<Vec<_>>())
            .finish()
    }
}

impl Worker {
    pub fn new() -> Self {
        Worker {
            handlers: HashMap::new(),
        }
    }

    pub fn register_handler_raw<F, R, J>(&mut self, name: J, handler: F)
    where
        F: Fn(Job<rmpv::Value>) -> R + 'static,
        R: Into<JobResult>,
        J: Into<String>,
    {
        self.handlers
            .insert(name.into(), Box::new(move |job| handler(job).into()));
    }

    pub fn register_handler<A, F, R, J>(&mut self, name: J, handler: F)
    where
        A: for<'de> Deserialize<'de>,
        F: Fn(Job<A>) -> R + 'static,
        R: Into<JobResult>,
        J: Into<String>,
    {
        let new_handler = move |job: Job<rmpv::Value>| {
            let job = job.map_argument(|v| rmpv::ext::from_value::<A>(v).unwrap());
            handler(job)
        };
        self.register_handler_raw(name.into(), new_handler);
    }

    pub fn register_handler_simple<A, F, R, J>(&mut self, name: J, handler: F)
    where
        A: for<'de> Deserialize<'de>,
        F: Fn(A) -> R + 'static,
        R: Into<JobResult>,
        J: Into<String>,
    {
        let new_handler = move |job: Job<rmpv::Value>| {
            let job = job
                .map_argument(|v| rmpv::ext::from_value::<A>(v).unwrap())
                .get_argument();
            handler(job)
        };
        self.register_handler_raw(name.into(), new_handler);
    }
}

#[cfg(test)]
mod test {
    use serde::Deserialize;

    use crate::models::JobResult;

    #[test]
    fn test_register_handler() {
        let mut worker = super::Worker::new();
        worker.register_handler_simple("Job1", handler1);
    }

    #[derive(Deserialize)]
    struct Sample1 {
        a: String,
    }

    #[derive(Deserialize)]
    struct Sample2 {
        b: i32,
    }

    fn handler1(data: Sample1) -> JobResult {
        todo!()
    }

    fn handler2(data: Sample2) -> JobResult {
        todo!()
    }
}
