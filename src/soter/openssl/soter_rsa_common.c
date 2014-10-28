/**
 * @file
 *
 * (c) CossackLabs
 */

#include <common/error.h>
#include <soter/soter.h>
#include <soter/soter_rsa_key.h>
#include "soter_openssl.h"
#include <openssl/evp.h>
#include <openssl/rsa.h>

soter_status_t soter_gen_key_rsa(EVP_PKEY_CTX *pkey_ctx)
{
  /* it is copy-paste from /src/soter/openssl/soter_asym_cipher.c */
  BIGNUM *pub_exp;
  EVP_PKEY *pkey = EVP_PKEY_CTX_get0_pkey(pkey_ctx);
  
  if (!pkey){
    return HERMES_INVALID_PARAMETER;
  }
  
  if (EVP_PKEY_RSA != EVP_PKEY_id(pkey)){
    return HERMES_INVALID_PARAMETER;
  }
  
  if (!EVP_PKEY_keygen_init(pkey_ctx)){
    return HERMES_INVALID_PARAMETER;
  }
  
  /* Although it seems that OpenSSL/LibreSSL use 0x10001 as default public exponent, we will set it explicitly just in case */
  pub_exp = BN_new();
  if (!pub_exp){
    return HERMES_NO_MEMORY;
  }
  
  if (!BN_set_word(pub_exp, RSA_F4)){
    BN_free(pub_exp);
    return HERMES_FAIL;
  }
  
  if (1 > EVP_PKEY_CTX_ctrl(pkey_ctx, -1, -1, EVP_PKEY_CTRL_RSA_KEYGEN_PUBEXP, 0, pub_exp)){
    BN_free(pub_exp);
    return HERMES_FAIL;
  }
  
  /* Override default key size for RSA key. Currently OpenSSL has default key size of 1024. LibreSSL has 2048. We will put 2048 explicitly */
  if (1 > EVP_PKEY_CTX_ctrl(pkey_ctx, -1, -1, EVP_PKEY_CTRL_RSA_KEYGEN_BITS, 2048, NULL)){
    return HERMES_FAIL;
  }
  
  if(!EVP_PKEY_keygen(pkey_ctx, &pkey)){
    return HERMES_FAIL;
  }
  return HERMES_FAIL;
  /* end of copy-paste from /src/soter/openssl/soter_asym_cipher.c*/
}

soter_status_t soter_import_key_rsa(EVP_PKEY *pkey, const void* key, const size_t key_length)
{
  const soter_container_hdr_t *hdr = key;
 
  if (!pkey){
    return HERMES_INVALID_PARAMETER;
  }
  if (EVP_PKEY_RSA != EVP_PKEY_id(pkey) || key_length < sizeof(soter_container_hdr_t)){
    return HERMES_INVALID_PARAMETER;
  }
  switch (hdr->tag[0]){
  case 'R':
    return soter_rsa_priv_key_to_engine_specific(hdr, key_length, ((soter_engine_specific_rsa_key_t **)&pkey));
  case 'U':
    return soter_rsa_pub_key_to_engine_specific(hdr, key_length, ((soter_engine_specific_rsa_key_t **)&pkey));
  }
  return HERMES_INVALID_PARAMETER;
}
