vps {
  image_name = "{{ .VpsImage }}"
}

bootstrap {
  ops_user        = "{{ .OpsUser }}"
  public_key_path = "{{ .SshPublicKeyPath }}"
}
