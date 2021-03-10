FROM alpine:3.4
RUN apk -U add ca-certificates
EXPOSE 9090
ADD bin/archer /bin/archer
CMD ["bin/archer", "serve", "--grpcPort", "9090"]