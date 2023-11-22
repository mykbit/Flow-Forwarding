1. **4 bytes - Sender ID**
2. **1 byte - Type of Transmission**
    - **Value 0:** Looking up the location of the server
    - **Value 1:** Client sending data to the server
    - **Value 2:** Server sending acknowledgement to the user
    - **Value 3:** Client requesting removal of its data from all the entities
3. **4 bytes - Destination ID**
