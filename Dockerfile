FROM gcr.io/distroless/static-debian11:nonroot
ENTRYPOINT ["/baton-sumo-logic"]
COPY baton-sumo-logic /