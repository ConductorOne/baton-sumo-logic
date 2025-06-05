While developing the connector, please fill out this form. This information is needed to write docs and to help other users set up the connector.

## Connector capabilities

1. What resources does the connector sync?

The Sumo Logic connector syncs the following resources:
- Users (both human accounts and service accounts)
- Roles

Note: Service account syncing can be optionally disabled using the `include-service-accounts` configuration parameter.

2. Can the connector provision any resources? If so, which ones? 

Yes, the connector supports provisioning capabilities for:
- User accounts (create and delete)
- Role assignments (granting and revoking role memberships to users)

## Connector credentials 

1. What credentials or information are needed to set up the connector? (For example, API key, client ID and secret, domain, etc.)

The following credentials and configuration are required:
- API Base URL (Required, defaults to US1 endpoint)
- API Access ID (Required)
- API Access Key (Required)
- Include Service Accounts flag (Optional, defaults to true)

2. For each item in the list above: 

   * How does a user create or look up that credential or info? Please include links to (non-gated) documentation, screenshots (of the UI or of gated docs), or a video of the process. 

   For API Base URL:
   - Determine your Sumo Logic deployment region
   - Use the corresponding API endpoint from these options:
     - AU: https://api.au.sumologic.com
     - CA: https://api.ca.sumologic.com
     - DE: https://api.de.sumologic.com
     - EU: https://api.eu.sumologic.com
     - FED: https://api.fed.sumologic.com
     - IN: https://api.in.sumologic.com
     - JP: https://api.jp.sumologic.com
     - KR: https://api.kr.sumologic.com
     - US1 (default): https://api.sumologic.com
     - US2: https://api.us2.sumologic.com

   For API Access ID and Key:
   You can obtain these credentials either through Personal Access Keys or Service Account Keys:

   Option 1 - Personal Access Keys:
   1. Log into Sumo Logic web interface
   2. Navigate to Preferences > Personal Access Keys
   3. Click "+ Add Access Key"
   4. Enter a name for the key
   5. Select appropriate scopes (permissions)
   6. Click Save to generate the key
   7. Copy both the Access ID and Access Key (they are only shown once)

   Option 2 - Service Account Keys:
   1. Log into Sumo Logic web interface
   2. Navigate to Administration > Security > Service Accounts
   3. Click "+ New" to create a new service account (if needed)
   4. Select the service account you want to use
   5. Click on "Add Access Key"
   6. Enter a name for the key
   7. Select appropriate scopes (permissions)
   8. Click Save to generate the key
   9. Copy both the Access ID and Access Key (they are only shown once)

   * Does the credential need any specific scopes or permissions? If so, list them here. 

   The API access key and user creating it need the following permissions:
    
   Base Requirements:
   - "Create Access Keys" role capability
   - "Manage Access Keys" capability
   - An Administrator role, or a custom role with the 'Create Access Keys' and 'Manage Access Keys' capabilities.

   For all operations (sync and provisioning):
   - Administrator role or role with "Manage Users and Roles" capability

   * If applicable: Is the list of scopes or permissions different to sync (read) versus provision (read-write)? If so, list the difference here. 

   Yes, the permissions differ between sync and provisioning operations:
   - For sync (read) operations: "View Users and Roles" capability is sufficient
   - For provisioning (read-write) operations: "Manage Users and Roles" capability is required
   - An Administrator role includes both capabilities and can perform all operations

   * What level of access or permissions does the user need in order to create the credentials? (For example, must be a super administrator, must have access to the admin console, etc.)  

   The user needs:
   - "Create Access Keys" role capability to create new access keys
   - "Manage Access Keys" capability to manage keys across the organization
   - Sufficient permissions matching the intended use of the key (read/write access to users and roles)

   **Additional Resources:**
   - [Sumo Logic Access Keys Documentation](https://help.sumologic.com/docs/manage/security/access-keys/)
   - [API Authentication and Endpoints](https://help.sumologic.com/docs/api/getting-started/)

   **Troubleshooting Tips:**
   - Ensure you have the necessary role capabilities before attempting to create an access key.
   - If you encounter issues with permissions, verify that your user role includes all required capabilities.
   - Remember to copy the Access ID and Access Key immediately after creation, as they are only displayed once.