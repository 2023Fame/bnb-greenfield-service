# Decenter-FileService

# 1. Introduction
Decenter FileService is a Go-based Web 2.0 service integral to the Decenter project. It serves as a dedicated storage solution for user-uploaded training data, pre-trained models, and the models processed by workers. Key features include:

- Upload and Download Support: Seamless integration allowing both uploads from users and downloads by the network's workers.
- Integration with BNB Greenfield: Data is securely stored in the BNB Greenfield, ensuring reliability and scalability.
-  Before storing the data in BNB, we will use AES encryption, and then when downloading the data, we will try to decrypt the data first and then send it to the user
- Open source processing: By maintaining transparency and openness in intermediate processing, we make every effort to ensure Decenter's decentralization

Through Decenter FileService, we strive to provide a robust storage solution while preserving the decentralized ethos of the Decenter project.
## [2. BNB Storage Client Guide](BnbStorageClient.md)

## [3. Deploy project to azure web app.](How_to_deploy.md)

