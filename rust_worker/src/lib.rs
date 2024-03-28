pub mod models;
mod pusher;
mod transport;
mod worker;

pub fn add(left: usize, right: usize) -> usize {
    left + right
}

pub(crate) fn gen_id() -> String {
    uuid::Uuid::new_v7(uuid::Timestamp::now(uuid::NoContext))
        .simple()
        .to_string()
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn it_works() {
        let result = add(2, 2);
        assert_eq!(result, 4);
    }
}
