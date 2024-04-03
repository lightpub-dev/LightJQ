use std::collections::HashMap;

use derive_getters::Getters;
use redis::{FromRedisValue, ToRedisArgs};
use serde::{Deserialize, Serialize};
use thiserror::Error;

#[derive(Debug, Error)]
pub enum JQError {
    #[error("Failed to encode data: {0}")]
    SerializationError(#[from] rmp_serde::encode::Error),
    #[error("Failed to decode data: {0}")]
    DeserializationError(#[from] rmp_serde::decode::Error),
    #[error("Redis error: {0}")]
    RedisError(#[from] redis::RedisError),
}

pub type Result<T> = std::result::Result<T, JQError>;

macro_rules! impl_redis {
    ($t:ty) => {
        impl ToRedisArgs for $t {
            fn write_redis_args<W>(&self, out: &mut W)
            where
                W: ?Sized + redis::RedisWrite,
            {
                let data = encode_data(self).unwrap();
                out.write_arg(data.as_slice())
            }
        }

        impl FromRedisValue for $t {
            fn from_redis_value(v: &redis::Value) -> redis::RedisResult<Self> {
                let data = match v {
                    redis::Value::Data(data) => data,
                    _ => redis::RedisResult::Err(
                        (redis::ErrorKind::TypeError, "unexpected type").into(),
                    )?,
                };
                decode_data(data).map_err(|e| match e {
                    JQError::DeserializationError(e) => {
                        (redis::ErrorKind::TypeError, "decode error", e.to_string()).into()
                    }
                    _ => panic!("unexpected error: {:?}", e),
                })
            }
        }
    };
}

#[derive(Debug, Serialize, Deserialize, Getters)]
pub struct Worker {
    id: String,
    #[serde(rename = "worker_name")]
    name: String,
    processes: i32,
}

#[derive(Debug, Serialize, Deserialize, Getters)]
pub struct JobRequest<A> {
    id: String,
    name: String,
    argument: A,
    priority: i32,
    max_retry: i32,
    keep_result: bool,
    timeout: i64,
}

#[derive(Debug, Serialize, Deserialize, Getters)]
pub struct Job<A> {
    id: String,
    name: String,
    argument: A,
    priority: i32,
    max_retry: i32,
    keep_result: bool,
    timeout: i64,
    registered_at: chrono::DateTime<chrono::Utc>,
}

impl<A> Job<A> {
    pub fn get_argument(self) -> A {
        self.argument
    }

    pub fn map_argument<B>(self, f: impl FnOnce(A) -> B) -> Job<B> {
        Job {
            id: self.id,
            name: self.name,
            argument: f(self.argument),
            priority: self.priority,
            max_retry: self.max_retry,
            keep_result: self.keep_result,
            timeout: self.timeout,
            registered_at: self.registered_at,
        }
    }
}

#[derive(Debug, Serialize, Deserialize)]
#[serde(tag = "type")]
pub enum JobResult {
    Success(JobResultSuccess),
    Failure(JobResultFailure),
}

#[derive(Debug, Serialize, Deserialize, Getters)]
pub struct JobResultSuccess {
    id: String,
    finished_at: chrono::DateTime<chrono::Utc>,
    result: rmpv::Value,
}

#[derive(Debug, Serialize, Deserialize, Getters)]
pub struct JobResultFailure {
    id: String,
    finished_at: chrono::DateTime<chrono::Utc>,
    message: String,
    reason: FailureReason,
    should_retry: bool,
    error: rmpv::Value,
}

#[derive(Debug, Serialize, Deserialize)]
pub enum FailureReason {
    #[serde(rename = "timeout")]
    Timeout,
    #[serde(rename = "other")]
    Other,
}

impl_redis!(Worker);
impl_redis!(JobResult);
impl_redis!(JobResultSuccess);
impl_redis!(JobResultFailure);
impl_redis!(FailureReason);

impl<A: Serialize> ToRedisArgs for JobRequest<A> {
    fn write_redis_args<W>(&self, out: &mut W)
    where
        W: ?Sized + redis::RedisWrite,
    {
        let data = encode_data(self).unwrap();
        out.write_arg(data.as_slice())
    }
}

impl<'de, A: Deserialize<'de>> FromRedisValue for Job<A> {
    fn from_redis_value(v: &redis::Value) -> redis::RedisResult<Self> {
        let data = match v.clone() {
            redis::Value::Data(data) => data,
            _ => redis::RedisResult::Err((redis::ErrorKind::TypeError, "unexpected type").into())?,
        };
        let decoded = decode_data(&data).map_err(move |e| match e {
            JQError::DeserializationError(e) => {
                (redis::ErrorKind::TypeError, "decode error", e.to_string()).into()
            }
            _ => panic!("unexpected error: {:?}", e),
        });
        decoded
    }
}

pub fn encode_data<T>(data: &T) -> Result<Vec<u8>>
where
    T: Serialize,
{
    rmp_serde::to_vec(data).map_err(|e| e.into())
}

pub fn decode_data<'a, T>(data: &'a [u8]) -> Result<T>
where
    T: Deserialize<'a>,
{
    rmp_serde::from_slice(data).map_err(|e| e.into())
}

pub trait JobRequestBuilderDefault {
    fn default_timeout(&self) -> chrono::Duration;
    fn default_max_retry(&self) -> i32;
}

#[derive(Debug)]
pub struct JobRequestBuilder<'a, T, A> {
    pusher: &'a T,

    id: Option<String>,
    name: Option<String>,
    argument: Option<A>,
    priority: Option<i32>,
    max_retry: Option<i32>,
    keep_result: Option<bool>,
    timeout: Option<chrono::Duration>,
}

impl<'a, T: JobRequestBuilderDefault, A> JobRequestBuilder<'a, T, A> {
    pub fn new(pusher: &'a T) -> Self {
        JobRequestBuilder {
            pusher,
            id: None,
            name: None,
            argument: None,
            priority: None,
            max_retry: None,
            keep_result: None,
            timeout: None,
        }
    }

    pub fn build(self) -> JobRequest<A> {
        JobRequest {
            id: self.id.expect("id not set"),
            name: self.name.expect("name not set"),
            argument: self.argument.expect("argument not set"),
            priority: self.priority.unwrap_or(0),
            max_retry: self.max_retry.unwrap_or(self.pusher.default_max_retry()),
            keep_result: self.keep_result.unwrap_or(false),
            timeout: self
                .timeout
                .unwrap_or(self.pusher.default_timeout())
                .num_milliseconds(),
        }
    }

    pub fn id(&mut self, id: impl Into<String>) -> &mut Self {
        self.id = Some(id.into());
        self
    }

    pub fn name(&mut self, name: impl Into<String>) -> &mut Self {
        self.name = Some(name.into());
        self
    }

    pub fn argument(&mut self, argument: A) -> &mut Self {
        self.argument = Some(argument);
        self
    }

    pub fn priority(&mut self, priority: i32) -> &mut Self {
        self.priority = Some(priority);
        self
    }

    pub fn max_retry(&mut self, max_retry: i32) -> &mut Self {
        self.max_retry = Some(max_retry);
        self
    }

    pub fn keep_result(&mut self, keep_result: bool) -> &mut Self {
        self.keep_result = Some(keep_result);
        self
    }

    pub fn timeout(&mut self, timeout: chrono::Duration) -> &mut Self {
        self.timeout = Some(timeout);
        self
    }
}
