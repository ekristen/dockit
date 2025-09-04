# Database

Dockit uses [GORM](http://gorm.io) for its database connections and uses [Snowflake](https://github.com/bwmarrin/snowflake) to generate all the primary keys. The primary database that has been developed against and tested is SQLite. However MySQL has been tested as well.

Snowflake is an algorithm from Twitter about generating `int64` integers based on time, node id, and sequential bits that allow for thousands of writes per millisecond. For most users of dockit this once matter and because of that, the default setting is to use a random number per instance on start up and log that ID out in the logs for tracking.

If you'd like to manually manage the `node-id` you can in your deployment, the only reason you would need to do this and would care to do this is if you are expecting, 5000+ writes per millisecond to the registry.

**Note:** this registry is mostly read heavy, writes are only performed when doing permissions.
