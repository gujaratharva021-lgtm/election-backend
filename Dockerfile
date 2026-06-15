FROM scratch
COPY election_backend /election_backend
EXPOSE 5000
CMD ["/election_backend"]